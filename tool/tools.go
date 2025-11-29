package tool

import "encoding/json"

func JsonEncode(agrs interface{}) string {
	json, _ := json.Marshal(agrs)
	return string(json)
}
