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
