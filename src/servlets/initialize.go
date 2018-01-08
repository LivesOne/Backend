package servlets

import (
	"servlets/common"
	"servlets/token"
)

func Init() {
	common.RedisPoolInit()
	common.UserDbInit()
	common.AssetDbInit()
	common.InitTxHistoryMongoDB()

	token.Init()
}
