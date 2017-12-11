package httpHandlers

import (
	"server"
	"servlets/accounts"
	"servlets/httpcfg"
)

func RegisterHandlers() {

	server.RegisterHandler(httpCfg.HELLO_WORLD, &helloWorldHandler{})
	server.RegisterHandler(httpCfg.ECHO_MSG, &echoMsgHandler{})

	// register accounts handlers
	accounts.RegisterHandlers()
}
