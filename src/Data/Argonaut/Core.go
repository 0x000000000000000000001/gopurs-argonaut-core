

import (
	"bytes"
	"encoding/json"
	"sort"
	"gopurs/output/gopurs_runtime"
)

func deepUnbox(v interface{}) interface{} {
	if val, ok := v.(gopurs_runtime.Value); ok {
		switch val.Type {
		case gopurs_runtime.TypeInt:
			return val.IntVal
		case gopurs_runtime.TypeFloat:
			return val.FloatVal()
		case gopurs_runtime.TypeString:
			if val.UnsafePtr != nil {
				return *(*string)(val.UnsafePtr)
			}
			return ""
		case gopurs_runtime.TypeBool:
			return val.IntVal != 0
		case gopurs_runtime.TypeArray:
			if val.UnsafePtr != nil {
				arr := *(*[]gopurs_runtime.Value)(val.UnsafePtr)
				res := make([]interface{}, len(arr))
				for i, x := range arr {
					res[i] = deepUnbox(x)
				}
				return res
			}
			return []interface{}{}
		case gopurs_runtime.TypeRecord, gopurs_runtime.TypeRecord0, gopurs_runtime.TypeRecord1, gopurs_runtime.TypeRecord2, gopurs_runtime.TypeRecord3, gopurs_runtime.TypeRecord4, gopurs_runtime.TypeRecord5:
			rec := gopurs_runtime.RecordToMap(val)
			res := make(map[string]interface{})
			for k, x := range rec {
				res[k] = deepUnbox(x)
			}
			return res
		case gopurs_runtime.TypeAny:
			if val.UnsafePtr == nil {
				return nil
			}
			if *(*any)(val.UnsafePtr) == nil {
				return nil
			}
			return deepUnbox(*(*any)(val.UnsafePtr))
		default:
			return nil
		}
	}
	if valSlice, ok := v.([]gopurs_runtime.Value); ok {
		res := make([]interface{}, len(valSlice))
		for i, x := range valSlice {
			res[i] = deepUnbox(x)
		}
		return res
	}
	if mapRaw, ok := v.(map[string]interface{}); ok {
		res := make(map[string]interface{})
		for k, x := range mapRaw {
			res[k] = deepUnbox(x)
		}
		return res
	}
	if arrRaw, ok := v.([]interface{}); ok {
		res := make([]interface{}, len(arrRaw))
		for i, x := range arrRaw {
			res[i] = deepUnbox(x)
		}
		return res
	}
	return v
}

func FromBoolean(b any) any {
	return b
}

func FromNumber(n any) any {
	return n
}

func FromString(s any) any {
	return s
}

func FromArray(a any) any {
	return a
}

func FromObject(o any) any {
	return o
}

func JsonNull() any {
	return nil
}

func Stringify(j any) string {
	b, _ := json.Marshal(deepUnbox(j))
	return string(b)
}

func StringifyWithIndent(i int, j any) string {
	spaces := ""
	if i > 0 {
		if i > 10 {
			i = 10
		}
		for k := 0; k < i; k++ {
			spaces += " "
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", spaces)
	enc.SetEscapeHTML(false)
	enc.Encode(deepUnbox(j))
	return buf.String()
}

func isArray(a any) bool {
	_, ok := deepUnbox(a).([]any)
	return ok
}

func _Compare(EQ any, GT any, LT any, a any, b any) any {
	a = deepUnbox(a)
	b = deepUnbox(b)
	if a == nil {
		if b == nil {
			return EQ
		}
		return LT
	}
	switch va := a.(type) {
	case bool:
		if vb, ok := b.(bool); ok {
			if va == vb {
				return EQ
			} else if !va {
				return LT
			}
			return GT
		}
		if b == nil {
			return GT
		}
		return LT
	case float64:
		if vb, ok := b.(float64); ok {
			if va == vb {
				return EQ
			} else if va < vb {
				return LT
			}
			return GT
		}
		if b == nil || isBool(b) {
			return GT
		}
		return LT
	case string:
		if vb, ok := b.(string); ok {
			if va == vb {
				return EQ
			} else if va < vb {
				return LT
			}
			return GT
		}
		if b == nil || isBool(b) || isFloat64(b) {
			return GT
		}
		return LT
	case []any:
		if vb, ok := b.([]any); ok {
			minLen := len(va)
			if len(vb) < minLen {
				minLen = len(vb)
			}
			for i := 0; i < minLen; i++ {
				ca := _Compare(EQ, GT, LT, va[i], vb[i])
				if ca != EQ {
					return ca
				}
			}
			if len(va) == len(vb) {
				return EQ
			} else if len(va) < len(vb) {
				return LT
			}
			return GT
		}
		if b == nil || isBool(b) || isFloat64(b) || isString(b) {
			return GT
		}
		return LT
	case map[string]any:
		if vb, ok := b.(map[string]any); ok {
			if len(va) < len(vb) {
				return LT
			} else if len(va) > len(vb) {
				return GT
			}
			akeys := make([]string, 0, len(va))
			bkeys := make([]string, 0, len(vb))
			for k := range va {
				akeys = append(akeys, k)
			}
			for k := range vb {
				bkeys = append(bkeys, k)
			}
			keysMap := make(map[string]bool)
			for _, k := range akeys {
				keysMap[k] = true
			}
			for _, k := range bkeys {
				keysMap[k] = true
			}
			keys := make([]string, 0, len(keysMap))
			for k := range keysMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				valA, okA := va[k]
				valB, okB := vb[k]
				if !okA {
					return LT
				} else if !okB {
					return GT
				}
				ck := _Compare(EQ, GT, LT, valA, valB)
				if ck != EQ {
					return ck
				}
			}
			return EQ
		}
		return GT
	default:
		panic("unknown JSON type in _Compare")
	}
}

func isBool(v any) bool {
	_, ok := deepUnbox(v).(bool)
	return ok
}

func isFloat64(v any) bool {
	_, ok := deepUnbox(v).(float64)
	return ok
}

func isString(v any) bool {
	_, ok := deepUnbox(v).(string)
	return ok
}

