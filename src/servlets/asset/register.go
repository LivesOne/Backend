package asset

import (
	"server"
	"servlets/constants"
)

func RegisterHandlers() {
	server.RegisterHandler(constants.ASSET_REWARD, &rewardHandler{})
	server.RegisterHandler(constants.ASSET_BALANCE, &balanceHandler{})
	server.RegisterHandler(constants.ASSET_TRANS_PREPARE, &transPrepareHandler{})
	server.RegisterHandler(constants.ASSET_TRANS_COMMIT, &transCommitHandler{})
	server.RegisterHandler(constants.ASSET_TRANS_RESULT, &transResultHandler{})
	server.RegisterHandler(constants.ASSET_TRANS_HISTORY, &transHistoryHandler{})
	server.RegisterHandler(constants.ASSET_LOCK_CREATE, &lockCreateHandler{})
	server.RegisterHandler(constants.ASSET_LOCK_REMOVE, &lockRemoveHandler{})
	server.RegisterHandler(constants.ASSET_LOGK_LIST, &lockListHandler{})
}
