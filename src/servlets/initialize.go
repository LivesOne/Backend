package servlets

import (
	"servlets/common"
	"servlets/token"
)

func Init() {
	token.Init()
	common.DbInit()

}
