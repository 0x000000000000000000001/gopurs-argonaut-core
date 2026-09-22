package Parser

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unsafe"
)

type parserObservation struct {
	value          any
	failure        any
	successes      int
	failures       int
	returnedMarker bool
}

func observeParser(input string) parserObservation {
	var observation parserObservation
	marker := &observation
	result := _JsonParser(func(value any) any {
		observation.failures++
		observation.failure = value
		return marker
	}, func(value any) any {
		observation.successes++
		observation.value = value
		return marker
	}, input)
	observation.returnedMarker = result == marker
	return observation
}

// Unlike reflect.DeepEqual or JSON re-encoding, this checks negative zero,
// the concrete FFI types, and nil versus empty containers.
func equalParserValue(got, want any) bool {
	switch expected := want.(type) {
	case nil:
		return got == nil
	case bool:
		value, ok := got.(bool)
		return ok && value == expected
	case float64:
		value, ok := got.(float64)
		return ok && math.Float64bits(value) == math.Float64bits(expected)
	case string:
		value, ok := got.(string)
		return ok && value == expected
	case []any:
		value, ok := got.([]any)
		if !ok || len(value) != len(expected) || (value == nil) != (expected == nil) {
			return false
		}
		for i := range expected {
			if !equalParserValue(value[i], expected[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		value, ok := got.(map[string]any)
		if !ok || len(value) != len(expected) || (value == nil) != (expected == nil) {
			return false
		}
		for key, item := range expected {
			actual, present := value[key]
			if !present || !equalParserValue(actual, item) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("unexpected reference type %T", want))
	}
}

func compareParser(input string) error {
	var expected any
	err := json.Unmarshal([]byte(input), &expected)
	actual := observeParser(input)
	if !actual.returnedMarker {
		return fmt.Errorf("did not return the callback result")
	}
	if err != nil {
		if actual.failures != 1 || actual.successes != 0 {
			return fmt.Errorf("invalid input: successes=%d failures=%d", actual.successes, actual.failures)
		}
		if actual.failure != err.Error() {
			return fmt.Errorf("error mismatch: got %#v, want %q", actual.failure, err.Error())
		}
		return nil
	}
	if actual.successes != 1 || actual.failures != 0 {
		return fmt.Errorf("valid input: successes=%d failures=%d, error=%v", actual.successes, actual.failures, actual.failure)
	}
	if !equalParserValue(actual.value, expected) {
		return fmt.Errorf("different values or concrete FFI types")
	}
	return nil
}

func directedParserCases() []string {
	cases := []string{
		`null`, `true`, `false`, `0`, `-0`, `-0.0`, `-0e10`, `123`, `-123`,
		`9007199254740993`, `18446744073709551616`, `1.7976931348623157e308`, `1.7976931348623159e308`,
		`5e-324`, `-5e-324`, `1e-10000`, `-1e-10000`, `2.2250738585072014e-308`, `1e10000`, `-1e10000`,
		`""`, `"plain"`, `"é漢💻"`, `"\u0000\b\f\n\r\t\\\/\""`,
		`"\ud83d\udcbb"`, `"\uD834\uDD1E"`, `"\ud800"`, `"\udfff"`, `"\ud800a"`,
		`"\ud800\ud800\udc00\udfff"`, `"\udc00\ud800"`, `"\uffff\ufffe"`,
		`[]`, `{}`, `[null,[],{},false,true,-0,"x"]`,
		`{"x":1,"x":2}`, `{"x":{"a":1},"x":{"b":2}}`, `{"x":1,"\u0078":null}`,
		`{"\ud800":1,"\ufffd":2}`, `{"__proto__":{"x":1},"constructor":[],"":false}`,
		" \t\n {\"x\": 1} \r\n", "", " \t\n", `+1`, `01`, `-01`, `.1`, `1.`, `1e`, `1e+`, `--1`,
		`NaN`, `Infinity`, `undefined`, `True`, `NULL`, `nil`, `nul`, `tru`, `fals`,
		`[1,]`, `[,1]`, `{,}`, `{"x":}`, `{"x":1,}`, `{"x" 1}`, `{1:2}`, `{'x':1}`,
		`null true`, `{}[]`, `1x`, `/*hi*/1`, `1//hi`, "\ufeffnull", "\vnull", "\fnull", "null\u00a0",
		`"\x20"`, `"\v"`, `"\u000g"`, `"\u000"`, `"\"`, `"unterminated`, "\"line\nfeed\"", "\"zero\x00byte\"",
	}
	for _, bytes := range [][]byte{{0xff}, {0x80}, {0xc0, 0xaf}, {0xe2, 0x82}, {0xed, 0xa0, 0x80}, {0xf4, 0x90, 0x80, 0x80}, {0xf0, 0x9f, 0x92, 0xbb}} {
		cases = append(cases, `"`+string(bytes)+`"`, `{"`+string(bytes)+`": "x"}`, string(bytes))
	}
	return cases
}

func TestParserDirectedDifferential(t *testing.T) {
	inputs := directedParserCases()
	for i, input := range inputs {
		if err := compareParser(input); err != nil {
			t.Fatalf("case %d %q: %v", i, input, err)
		}
	}
	t.Logf("%d directed inputs match encoding/json", len(inputs))
}

func generatedParserValue(random *rand.Rand, depth int) any {
	choices := 6
	if depth == 0 {
		choices = 4
	}
	switch random.Intn(choices) {
	case 0:
		return nil
	case 1:
		return random.Intn(2) == 1
	case 2:
		return math.Ldexp(random.Float64()*2-1, random.Intn(1800)-900)
	case 3:
		values := []string{"", "ordinary", "é漢💻", "\n\r\t\x00\"\\", "\u2028\u2029", string([]byte{0xff, 0x80})}
		return values[random.Intn(len(values))]
	case 4:
		array := make([]any, random.Intn(7))
		for i := range array {
			array[i] = generatedParserValue(random, depth-1)
		}
		return array
	default:
		object := make(map[string]any)
		for n := random.Intn(7); n > 0; n-- {
			object[fmt.Sprintf("key%d_é", n)] = generatedParserValue(random, depth-1)
		}
		return object
	}
}

func TestParserGeneratedDifferential(t *testing.T) {
	random := rand.New(rand.NewSource(0x41c07))
	count := 0
	check := func(input string) {
		t.Helper()
		count++
		if err := compareParser(input); err != nil {
			t.Fatalf("generated case %d %q: %v", count, input, err)
		}
	}
	for i := 0; i < 750; i++ {
		bytes, err := json.Marshal(generatedParserValue(random, 4))
		if err != nil {
			t.Fatal(err)
		}
		input := string(bytes)
		check(input)
		check(input[:random.Intn(len(input)+1)])
		position := random.Intn(len(input))
		check(input[:position] + input[position+1:])
		check(input[:position] + string([]byte{byte(random.Intn(256))}) + input[position:])
		mutated := append([]byte(nil), bytes...)
		mutated[position] = byte(random.Intn(256))
		check(string(mutated))
		check(input + string([]byte{byte(random.Intn(256))}))
	}
	for _, input := range directedParserCases() {
		for end := 0; end < len(input); end++ {
			check(input[:end])
		}
	}
	t.Logf("%d generated, mutated and truncated inputs match encoding/json", count)
}

func TestParserDepthLimit(t *testing.T) {
	for _, depth := range []int{9999, 10000, 10001} {
		for _, shape := range [][2]string{{"[", "]"}, {`{"x":`, "}"}} {
			input := strings.Repeat(shape[0], depth) + "null" + strings.Repeat(shape[1], depth)
			if err := compareParser(input); err != nil {
				t.Fatalf("depth %d shape %q: %v", depth, shape[0], err)
			}
			if err := compareParser(input[:len(input)-1]); err != nil {
				t.Fatalf("truncated depth %d shape %q: %v", depth, shape[0], err)
			}
		}
	}
}

func TestParserDoesNotRetainInput(t *testing.T) {
	// Keeping one tiny string from a parsed tree must not retain the full
	// source document. Pointer-range checks avoid flaky heap/GC thresholds.
	input := strings.Clone(`{"key":"value","escaped\u006bey":"escaped\u0076alue","array":["inside"]}` + strings.Repeat(" ", 1<<20))
	actual := observeParser(input)
	if actual.successes != 1 || actual.failures != 0 {
		t.Fatal("retention fixture did not parse")
	}
	start := uintptr(unsafe.Pointer(unsafe.StringData(input)))
	end := start + uintptr(len(input))
	var walk func(any)
	check := func(value string) {
		if value == "" {
			return
		}
		pointer := uintptr(unsafe.Pointer(unsafe.StringData(value)))
		if pointer >= start && pointer < end {
			t.Errorf("parsed string %q retains the original input", value)
		}
	}
	walk = func(value any) {
		switch item := value.(type) {
		case string:
			check(item)
		case []any:
			for _, value := range item {
				walk(value)
			}
		case map[string]any:
			for key, value := range item {
				check(key)
				walk(value)
			}
		}
	}
	walk(actual.value)
	runtime.KeepAlive(input)
	runtime.KeepAlive(actual)
}

func TestParserConcurrentCalls(t *testing.T) {
	inputs := directedParserCases()
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for pass := 0; pass < 5; pass++ {
				for _, input := range inputs {
					if err := compareParser(input); err != nil {
						t.Errorf("concurrent input %q: %v", input, err)
						return
					}
				}
			}
		}()
	}
	workers.Wait()
}

func TestParserCallbackPanicsPropagate(t *testing.T) {
	for _, input := range []string{`{"valid":true}`, `invalid`} {
		func() {
			marker := new(int)
			calls := 0
			defer func() {
				if got := recover(); got != marker {
					t.Errorf("callback panic changed: %#v", got)
				}
				if calls != 1 {
					t.Errorf("callback called %d times", calls)
				}
			}()
			callback := func(any) any { calls++; panic(marker) }
			_JsonParser(callback, callback, input)
		}()
	}
}

func normalizedParserValue(value any) any {
	switch item := value.(type) {
	case nil:
		return []any{"null"}
	case bool:
		return []any{"boolean", item}
	case string:
		return []any{"string", item}
	case float64:
		return []any{"number", fmt.Sprintf("%016x", math.Float64bits(item))}
	case []any:
		result := make([]any, len(item))
		for i, value := range item {
			result[i] = normalizedParserValue(value)
		}
		return []any{"array", result}
	case map[string]any:
		result := make(map[string]any, len(item))
		for key, value := range item {
			result[key] = normalizedParserValue(value)
		}
		return []any{"object", result}
	default:
		panic(fmt.Sprintf("unexpected parser type %T", value))
	}
}

func TestParserJavaScriptCommonCases(t *testing.T) {
	bytes, err := os.ReadFile("js-oracle.json")
	if os.IsNotExist(err) {
		t.Skip("JS oracle supplied by node --test test/parser-native.test.mjs")
	}
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input    string
		Expected any
	}
	if err := json.Unmarshal(bytes, &cases); err != nil {
		t.Fatal(err)
	}
	for i, item := range cases {
		actual := observeParser(item.Input)
		if actual.successes != 1 || actual.failures != 0 {
			t.Fatalf("JS case %d %s: %v", i, strconv.Quote(item.Input), actual.failure)
		}
		if !reflect.DeepEqual(normalizedParserValue(actual.value), item.Expected) {
			t.Fatalf("JS case %d %s differs", i, strconv.Quote(item.Input))
		}
	}
	t.Logf("%d valid shared-contract inputs match JSON.parse", len(cases))
}

var parserBenchmarkSink any

func stdlibJSONParser(fail func(any) any, succ func(any) any, input string) any {
	var result any
	if err := json.Unmarshal([]byte(input), &result); err != nil {
		return fail(err.Error())
	}
	return succ(result)
}

// Run separately from correctness tests and real-corpus measurements:
// GOMAXPROCS=1 GOGC=100 ARGONAUT_PARSER_BENCH=1 node --test test/parser-native.test.mjs
func BenchmarkJSONParser(b *testing.B) {
	encode := func(value any) string {
		bytes, err := json.Marshal(value)
		if err != nil {
			b.Fatal(err)
		}
		return string(bytes)
	}
	numbers := make([]float64, 1024)
	objects := make([]any, 256)
	escaped := make([]string, 1024)
	slashes := make([]string, 1024)
	surrogates := make([]string, 1024)
	for i := range numbers {
		numbers[i] = float64(i-500) / 3.7
		escaped[i] = "a\nb\t\"c"
		slashes[i] = `"https:\/\/example.test\/path"`
		surrogates[i] = `"\ud83d\udcbb"`
	}
	for i := range objects {
		objects[i] = map[string]any{"id": float64(i), "name": "item", "enabled": true, "value": nil}
	}
	cases := []struct{ name, input string }{
		{"small-scalar", `"hello"`},
		{"ascii-string", encode(strings.Repeat("ASCIIabcdefgh01234567", 2048))},
		{"unicode-string", encode(strings.Repeat("é雪💻", 4096))},
		{"escaped-string", encode(strings.Repeat("line\n\tquote\"slash\\\x00", 2048))},
		{"escaped-array", encode(escaped)},
		{"slash-array", "[" + strings.Join(slashes, ",") + "]"},
		{"surrogate-array", "[" + strings.Join(surrogates, ",") + "]"},
		{"number-array", encode(numbers)},
		{"object-array", encode(objects)},
	}
	parsers := []struct {
		name  string
		parse func(func(any) any, func(any) any, string) any
	}{{"stdlib", stdlibJSONParser}, {"native", _JsonParser}}
	fail := func(value any) any { panic(value) }
	success := func(value any) any { return value }
	for _, item := range cases {
		b.Run(item.name, func(b *testing.B) {
			for _, parser := range parsers {
				b.Run(parser.name, func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(len(item.input)))
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						parserBenchmarkSink = parser.parse(fail, success, item.input)
					}
				})
			}
		})
	}
}
