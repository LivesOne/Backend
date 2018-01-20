package accounts

import (
	"encoding/json"
	"servlets/constants"
	"servlets/token"
	"strconv"
	"time"
	"utils"
)

func DecryptSecret(secret string, key string, iv string, instance interface{}) constants.Error {
	dataStr, err := utils.AesDecrypt(secret, key, iv)
	if err != nil {
		return constants.RC_PARAM_ERR
	}
	if err := json.Unmarshal([]byte(dataStr), instance); err != nil {
		return constants.RC_PARAM_ERR
	}
	return constants.RC_OK
}

func TokenErr2RcErr(tokenErr int) constants.Error {
	switch tokenErr {
	case constants.ERR_INT_OK:
		return constants.RC_OK
	case constants.ERR_INT_TK_DB:
		return constants.RC_PARAM_ERR
	case constants.ERR_INT_TK_DUPLICATE:
		return constants.RC_PARAM_ERR
	case constants.ERR_INT_TK_NOTEXISTS:
		return constants.RC_PARAM_ERR
	default:
		return constants.RC_SYSTEM_ERR
	}
}


// 生成 request 所需的 signature
func GenerateSig(hash string) (string, constants.Error) {
	_, key, _, err := token.GetAll(hash)
	switch err {
	case constants.ERR_INT_OK:
		break
	case constants.ERR_INT_TK_DB:
		return "", constants.RC_PARAM_ERR
	case constants.ERR_INT_TK_DUPLICATE:
		return "", constants.RC_PARAM_ERR
	case constants.ERR_INT_TK_NOTEXISTS:
		return "", constants.RC_PARAM_ERR
	default:
		return "", constants.RC_SYSTEM_ERR
	}
	timestamp := GetTimestamp()
	in := key + timestamp
	sig := utils.Sha256(in)
	return sig, constants.RC_OK
}

// 获取13位时间戳
func GetTimestamp() string {
	now := time.Now()
	timestamp := now.UnixNano() / 1000000
	timestampString := strconv.FormatInt(timestamp, 10)
	return timestampString
}
