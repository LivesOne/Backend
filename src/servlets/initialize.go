package servlets

import (
	"servlets/common"
	"servlets/token"
)

func Init() {
	token.Init()
	common.UserDbInit()
	common.AssetDbInit()
	common.InitTxHistoryMongoDB()
}
