package message

import (
	"server"
	"servlets/constants"
)

func RegisterHandlers() {

	server.RegisterHandler(constants.MESSAGE_LIST, new(messageListHandler))
}
