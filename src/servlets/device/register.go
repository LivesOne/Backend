package device

import (
	"server"
	"servlets/constants"
)

func RegisterHandlers() {
	server.RegisterHandler(constants.DEVICE_BIND_DEVICE, new(deviceBindHandler))
	server.RegisterHandler(constants.DEVICE_UNBIND_DEVICE, new(deviceUnBindHandler))
	server.RegisterHandler(constants.DEVICE_DEVICE_INFO, new(deviceInfoHandler))
	server.RegisterHandler(constants.DEVICE_DEVICE_LIST, new(deviceListHandler))
}
