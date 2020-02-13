package servlets

import (
	"server"
	"servlets/accounts"
	"servlets/asset"
	"servlets/config"
	"servlets/constants"
	"servlets/contacts"
	"servlets/device"
	"servlets/message"
)

func RegisterHandlers() {

	server.RegisterHandler(constants.HelloWorld, &helloWorldHandler{})
	server.RegisterHandler(constants.Echo, &echoMsgHandler{})

	// register accounts handlers
	accounts.RegisterHandlers()
	contacts.RegisterHandlers()
	asset.RegisterHandlers()
	device.RegisterHandlers()
	config.RegisterHandlers()
	message.RegisterHandlers()
}
