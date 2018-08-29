package servlets

import (
	"servlets/common"
	"servlets/token"
)

func Init() {
	common.RedisPoolInit()
	common.UserDbInit()
	common.AssetDbInit()
	common.ConfigDbInit()
	common.InitTxHistoryMongoDB()
	common.InitMinerRMongoDB()
	common.InitTradeMongoDB()
	token.Init()
}
