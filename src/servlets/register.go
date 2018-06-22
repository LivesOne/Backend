package servlets

import (
	"server"
	"servlets/accounts"
	"servlets/constants"
	"servlets/device"
	"servlets/asset"
)

func RegisterHandlers() {

	server.RegisterHandler(constants.HelloWorld, &helloWorldHandler{})
	server.RegisterHandler(constants.Echo, &echoMsgHandler{})

	// register accounts handlers
	accounts.RegisterHandlers()
	asset.RegisterHandlers()
	device.RegisterHandlers()
}
