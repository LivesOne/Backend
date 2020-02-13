package accounts

import (
	"encoding/json"
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"servlets/constants"
	"servlets/rpc"
	"strconv"
	"time"
	"utils"
	"utils/logger"
)

func DecryptSecret(secret string, key string, iv string, instance interface{}) constants.Error {
	dataStr, err := utils.AesDecrypt(secret, key, iv)
	if err != nil {
		return constants.RC_PARAM_ERR
	}
	logger.Info("Decrypt Secret str ", dataStr)
	if err := json.Unmarshal([]byte(dataStr), instance); err != nil {
		return constants.RC_PARAM_ERR
	}
	return constants.RC_OK
}

// 生成 request 所需的 signature
func GenerateSig(hash string) (string, constants.Error) {
	_, key, _, err := rpc.GetTokenInfo(hash)
	switch err {
	case microuser.ResCode_OK:
		break
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
