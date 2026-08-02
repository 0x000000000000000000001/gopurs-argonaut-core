package Parser

import (
	"encoding/json"
)

func _JsonParser(fail func(any) any, succ func(any) any, str string) any {
	var result any
	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		return fail(err.Error())
	}
	return succ(result)
}
