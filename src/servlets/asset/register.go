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
	server.RegisterHandler(constants.ASSET_LOCK_UPGRADE, &lockUpgradeHandler{})
	server.RegisterHandler(constants.ASSET_WITHDRAWAL_QUOTA, &withdrawQuotaHandler{})
	server.RegisterHandler(constants.ASSET_ETH_TRANS_PREPARE, &ethTransPrepareHandler{})
	server.RegisterHandler(constants.ASSET_ETH_TRANS_COMMIT, &ethTransCommitHandler{})
	server.RegisterHandler(constants.ASSET_ETH_TRANS_RESULT, &ethtransResultHandler{})
	server.RegisterHandler(constants.ASSET_ETH_TRANS_HISTORY, &ethtransHistoryHandler{})
	server.RegisterHandler(constants.ASSET_WITHDRAWAL_LIST, &withdrawListHandler{})
	server.RegisterHandler(constants.ASSET_WITHDRAWAL_REQUEST, &withdrawRequestHandler{})
	server.RegisterHandler(constants.ASSET_WITHDRAWAL_CARD_LIST, &withdrawCardListHandler{})
	server.RegisterHandler(constants.ASSET_WITHDRAWAL_CARD_USE, &withdrawCardUseHandler{})
	server.RegisterHandler(constants.ASSET_WITHDRAWAL_CARD_USE_LIST, &withdrawCardUseListHandler{})


}
