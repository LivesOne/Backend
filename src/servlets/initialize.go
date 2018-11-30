package servlets

import (
	"servlets/common"
)

func Init() {
	common.RedisPoolInit()
	//common.UserDbInit()
	common.AssetDbInit()
	common.ConfigDbInit()
	common.InitTxHistoryMongoDB()
	common.InitMinerRMongoDB()
	common.InitTradeMongoDB()
	common.InitContactsMongoDB()
	common.InitMsgMongoDB()
}
