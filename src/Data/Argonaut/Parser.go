package Parser

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func _JsonParser(fail func(any) any, succ func(any) any, str string) any {
	result, err := argonautParseJSON(str)
	if err != nil {
		return fail(err.Error())
	}
	return succ(result)
}

type argonautJSONParser struct {
	text string
	pos  int
}

// Parse the existing Json representation in one pass over the input string.
// Keep encoding/json as the authority for malformed inputs and error messages.
// Extracted strings own their storage, so a small result does not retain the
// complete source document. Unicode replacement follows encoding/json's rules.
func argonautParseJSON(text string) (any, error) {
	p := argonautJSONParser{text: text}
	value, ok := p.value(0)
	p.space()
	if ok && p.pos == len(text) {
		return value, nil
	}
	var fallback any
	err := json.Unmarshal([]byte(text), &fallback)
	return fallback, err
}

func (p *argonautJSONParser) space() {
	for p.pos < len(p.text) {
		switch p.text[p.pos] {
		case ' ', '\n', '\r', '\t':
			p.pos++
		default:
			return
		}
	}
}

func (p *argonautJSONParser) value(depth int) (any, bool) {
	p.space()
	if p.pos >= len(p.text) || depth > 10000 {
		return nil, false
	}
	switch p.text[p.pos] {
	case '"':
		return p.string()
	case '{':
		return p.object(depth + 1)
	case '[':
		return p.array(depth + 1)
	case 't':
		if strings.HasPrefix(p.text[p.pos:], "true") {
			p.pos += 4
			return true, true
		}
	case 'f':
		if strings.HasPrefix(p.text[p.pos:], "false") {
			p.pos += 5
			return false, true
		}
	case 'n':
		if strings.HasPrefix(p.text[p.pos:], "null") {
			p.pos += 4
			return nil, true
		}
	default:
		return p.number()
	}
	return nil, false
}

func (p *argonautJSONParser) string() (string, bool) {
	if p.pos >= len(p.text) || p.text[p.pos] != '"' {
		return "", false
	}
	start := p.pos
	p.pos++
	escaped, nonASCII := false, false
	for p.pos < len(p.text) {
		c := p.text[p.pos]
		p.pos++
		if c == '"' {
			if escaped || (nonASCII && !utf8.ValidString(p.text[start+1:p.pos-1])) {
				return argonautUnquoteJSON(p.text[start+1 : p.pos-1])
			}
			return strings.Clone(p.text[start+1 : p.pos-1]), true
		}
		if c < 0x20 {
			return "", false
		}
		if c >= utf8.RuneSelf {
			nonASCII = true
		}
		if c == '\\' {
			escaped = true
			if p.pos >= len(p.text) {
				return "", false
			}
			p.pos++
		}
	}
	return "", false
}

// Decode JSON escapes without starting a complete decoder for each string.
// The builder owns the output; invalid UTF-8 and unpaired UTF-16 surrogates
// become U+FFFD, as they do in encoding/json. Invalid JSON uses the full fallback.
func argonautUnquoteJSON(text string) (string, bool) {
	var output strings.Builder
	output.Grow(len(text))
	for i := 0; i < len(text); {
		start := i
		for i < len(text) && text[i] >= 0x20 && text[i] < utf8.RuneSelf && text[i] != '\\' {
			i++
		}
		output.WriteString(text[start:i])
		if i == len(text) {
			break
		}
		c := text[i]
		if c < 0x20 {
			return "", false
		}
		if c >= utf8.RuneSelf {
			r, size := utf8.DecodeRuneInString(text[i:])
			output.WriteRune(r)
			i += size
			continue
		}
		i++ // backslash
		if i == len(text) {
			return "", false
		}
		escape := text[i]
		i++
		switch escape {
		case '"', '\\', '/':
			output.WriteByte(escape)
		case 'b':
			output.WriteByte('\b')
		case 'f':
			output.WriteByte('\f')
		case 'n':
			output.WriteByte('\n')
		case 'r':
			output.WriteByte('\r')
		case 't':
			output.WriteByte('\t')
		case 'u':
			r, ok := argonautJSONHex4(text, i)
			if !ok {
				return "", false
			}
			i += 4
			if r >= 0xd800 && r <= 0xdbff {
				if i+6 <= len(text) && text[i] == '\\' && text[i+1] == 'u' {
					low, valid := argonautJSONHex4(text, i+2)
					if valid && low >= 0xdc00 && low <= 0xdfff {
						r = utf16.DecodeRune(r, low)
						i += 6
					} else {
						r = utf8.RuneError
					}
				} else {
					r = utf8.RuneError
				}
			} else if r >= 0xdc00 && r <= 0xdfff {
				r = utf8.RuneError
			}
			output.WriteRune(r)
		default:
			return "", false
		}
	}
	return output.String(), true
}

func argonautJSONHex4(text string, pos int) (rune, bool) {
	if len(text)-pos < 4 {
		return 0, false
	}
	var value rune
	for _, c := range []byte(text[pos : pos+4]) {
		value <<= 4
		switch {
		case c >= '0' && c <= '9':
			value |= rune(c - '0')
		case c >= 'a' && c <= 'f':
			value |= rune(c-'a') + 10
		case c >= 'A' && c <= 'F':
			value |= rune(c-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

func (p *argonautJSONParser) number() (any, bool) {
	start := p.pos
	if p.pos < len(p.text) && p.text[p.pos] == '-' {
		p.pos++
	}
	if p.pos >= len(p.text) {
		return nil, false
	}
	if p.text[p.pos] == '0' {
		p.pos++
	} else {
		if p.text[p.pos] < '1' || p.text[p.pos] > '9' {
			return nil, false
		}
		for p.pos < len(p.text) && p.text[p.pos] >= '0' && p.text[p.pos] <= '9' {
			p.pos++
		}
	}
	if p.pos < len(p.text) && p.text[p.pos] == '.' {
		p.pos++
		digits := p.pos
		for p.pos < len(p.text) && p.text[p.pos] >= '0' && p.text[p.pos] <= '9' {
			p.pos++
		}
		if p.pos == digits {
			return nil, false
		}
	}
	if p.pos < len(p.text) && (p.text[p.pos] == 'e' || p.text[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.text) && (p.text[p.pos] == '+' || p.text[p.pos] == '-') {
			p.pos++
		}
		digits := p.pos
		for p.pos < len(p.text) && p.text[p.pos] >= '0' && p.text[p.pos] <= '9' {
			p.pos++
		}
		if p.pos == digits {
			return nil, false
		}
	}
	n, err := strconv.ParseFloat(p.text[start:p.pos], 64)
	return n, err == nil
}

func (p *argonautJSONParser) object(depth int) (any, bool) {
	if depth > 10000 {
		return nil, false
	}
	p.pos++
	p.space()
	obj := make(map[string]any)
	if p.pos < len(p.text) && p.text[p.pos] == '}' {
		p.pos++
		return obj, true
	}
	for {
		key, ok := p.string()
		if !ok {
			return nil, false
		}
		p.space()
		if p.pos >= len(p.text) || p.text[p.pos] != ':' {
			return nil, false
		}
		p.pos++
		value, ok := p.value(depth)
		if !ok {
			return nil, false
		}
		obj[key] = value
		p.space()
		if p.pos >= len(p.text) {
			return nil, false
		}
		c := p.text[p.pos]
		p.pos++
		if c == '}' {
			return obj, true
		}
		if c != ',' {
			return nil, false
		}
		p.space()
	}
}

func (p *argonautJSONParser) array(depth int) (any, bool) {
	if depth > 10000 {
		return nil, false
	}
	p.pos++
	p.space()
	values := make([]any, 0)
	if p.pos < len(p.text) && p.text[p.pos] == ']' {
		p.pos++
		return values, true
	}
	for {
		value, ok := p.value(depth)
		if !ok {
			return nil, false
		}
		values = append(values, value)
		p.space()
		if p.pos >= len(p.text) {
			return nil, false
		}
		c := p.text[p.pos]
		p.pos++
		if c == ']' {
			return values, true
		}
		if c != ',' {
			return nil, false
		}
	}
}
