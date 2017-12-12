package servlets

import (
	"server"
	"servlets/accounts"
	"servlets/constants"
)

func RegisterHandlers() {

	server.RegisterHandler(constants.HelloWorld, &helloWorldHandler{})
	server.RegisterHandler(constants.Echo, &echoMsgHandler{})

	// register accounts handlers
	accounts.RegisterHandlers()
}
