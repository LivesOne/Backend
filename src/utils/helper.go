package utils

import (
	"encoding/json"
	"utils/logger"
)

// ToJSONIndent is a helper method that converts the object to human readable format
func ToJSONIndent(v interface{}) string {

	ret, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		return string(ret)
	} else {
		logger.Info("MarshalIndent object to json failed", v)
		return ""
	}
}


func ToJSON(v interface{}) string {

	ret, err := json.Marshal(v)
	if err != nil {
		logger.Info("MarshalIndent object to json failed", v)
		return ""
	}
	return string(ret)
}


func FromJson(jsonStr string,v interface{})error{
	return json.Unmarshal([]byte(jsonStr),v)
}