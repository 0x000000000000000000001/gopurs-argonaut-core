package Parser

import (
	"encoding/json"
	"gopurs/output/gopurs_runtime"
)

func _JsonParser(fail interface{}, succ interface{}, s interface{}) interface{} {
	str := s.(gopurs_runtime.Value).StrVal()
	var result interface{}
	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		return gopurs_runtime.Apply(fail.(gopurs_runtime.Value), gopurs_runtime.Str(err.Error()))
	}
	// We return success and we box result in gopurs_runtime.Any
	// because `_CaseJson` and others expect to receive it as `interface{}`.
	// When Purescript passes it around, it will be a Value of type TypeAny.
	// However, `FromBoolean` etc. return native types directly because they return `interface{}`.
	// Wait, if FromBoolean returns a bool, it's boxed implicitly by gopurs as Any?
	// No, if FromBoolean returns bool, the gopurs wrapper does gopurs_runtime.Box(res),
	// which automatically converts bool to Value{TypeBool, ...} or Value{TypeAny, ...}.
	// We are returning to Purescript space, so the caller of `_JsonParser` will take our `interface{}`
	// and box it using `gopurs_runtime.Box(res)`.
	// So we can just return `result` directly! 
	return gopurs_runtime.Apply(succ.(gopurs_runtime.Value), gopurs_runtime.Box(result))
}
