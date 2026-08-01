

import (
	"bytes"
	"encoding/json"
	"gopurs/output/gopurs_runtime"
	"sort"
)

func FromBoolean(b interface{}) interface{} {
	return b
}

func FromNumber(n interface{}) interface{} {
	return n
}

func FromString(s interface{}) interface{} {
	return s
}

func FromArray(a interface{}) interface{} {
	return a
}

func FromObject(o interface{}) interface{} {
	return o
}

func JsonNull() interface{} {
	return nil
}

func Stringify(j interface{}) interface{} {
	b, _ := json.Marshal(j)
	return gopurs_runtime.Str(string(b))
}

func StringifyWithIndent(i interface{}, j interface{}) interface{} {
	indent := int(i.(gopurs_runtime.Value).IntVal)
	spaces := ""
	if indent > 10 {
		indent = 10
	}
	if indent > 0 {
		for k := 0; k < indent; k++ {
			spaces += " "
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", spaces)
	enc.SetEscapeHTML(false)
	enc.Encode(j)
	return gopurs_runtime.Str(buf.String())
}

func isArray(a interface{}) bool {
	_, ok := a.([]interface{})
	return ok
}

func _CaseJson(onNull interface{}, onBool interface{}, onNum interface{}, onStr interface{}, onArr interface{}, onObj interface{}, j interface{}) interface{} {
	if j == nil {
		return gopurs_runtime.Apply(onNull.(gopurs_runtime.Value), gopurs_runtime.Value{})
	}
	switch v := j.(type) {
	case bool:
		return gopurs_runtime.Apply(onBool.(gopurs_runtime.Value), gopurs_runtime.Any(v))
	case float64:
		return gopurs_runtime.Apply(onNum.(gopurs_runtime.Value), gopurs_runtime.Any(v))
	case int64:
		return gopurs_runtime.Apply(onNum.(gopurs_runtime.Value), gopurs_runtime.Any(float64(v)))
	case string:
		return gopurs_runtime.Apply(onStr.(gopurs_runtime.Value), gopurs_runtime.Any(v))
	case []interface{}:
		return gopurs_runtime.Apply(onArr.(gopurs_runtime.Value), gopurs_runtime.Any(v))
	case map[string]interface{}:
		return gopurs_runtime.Apply(onObj.(gopurs_runtime.Value), gopurs_runtime.Any(v))
	default:
		panic("unknown JSON type")
	}
}

func _Compare(EQ interface{}, GT interface{}, LT interface{}, a interface{}, b interface{}) interface{} {
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
	case []interface{}:
		if vb, ok := b.([]interface{}); ok {
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
	case map[string]interface{}:
		if vb, ok := b.(map[string]interface{}); ok {
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

func isBool(v interface{}) bool {
	_, ok := v.(bool)
	return ok
}

func isFloat64(v interface{}) bool {
	_, ok := v.(float64)
	return ok
}

func isString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

