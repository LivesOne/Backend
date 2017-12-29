package asset

import (
	"server"
	"servlets/constants"
)

func RegisterHandlers() {
	server.RegisterHandler(constants.ASSET_REWARD, &rewardHandler{})

}
