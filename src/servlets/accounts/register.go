package accounts

import (
	"server"
	"servlets/constants"
)

func RegisterHandlers() {
	server.RegisterHandler(constants.ACCOUNT_REGISTER, &registerUserHandler{})
	server.RegisterHandler(constants.ACCOUNT_REGISTER_UPVCODE, &registerUpVcodeHandler{})
	server.RegisterHandler(constants.ACCOUNT_LOGIN, &loginHandler{})
	server.RegisterHandler(constants.ACCOUNT_AUTOLOGIN, &autoLoginHandler{})
	server.RegisterHandler(constants.ACCOUNT_LOGOUT, &logoutHandler{})
	server.RegisterHandler(constants.ACCOUNT_GET_IMG_VCODE, &getImgVCodeHandler{})
	server.RegisterHandler(constants.ACCOUNT_SEND_VCODE, &sendVCodeHandler{})
	//server.RegisterHandler(constants.ACCOUNT_CHECK_VCODE, &checkVCodeHandler{})
	server.RegisterHandler(constants.ACCOUNT_MODIFY_PWD, &modifyPwdHandler{})
	server.RegisterHandler(constants.ACCOUNT_RESET_PWD, &resetPwdHandler{})
	server.RegisterHandler(constants.ACCOUNT_RESET_TX_PWD, &setTxPwdHandler{})
	server.RegisterHandler(constants.ACCOUNT_BIND_PHONE, &bindPhoneHandler{})
	server.RegisterHandler(constants.ACCOUNT_BIND_EMAIL, &bindEMailHandler{})
	server.RegisterHandler(constants.ACCOUNT_GET_PROFILE, &getProfileHandler{})
	server.RegisterHandler(constants.ACCOUNT_MODIFY_PROFILE, &modifyUserProfileHandler{})
	server.RegisterHandler(constants.ACCOUNT_CHECK_ACCOUNT, &checkAccountHandler{})
	server.RegisterHandler(constants.CHECK_WITH_VCODE, &checkWithVcodeHandler{})
	//server.RegisterHandler(constants.ACCOUNT_SET_STATUS, &setStatusHandler{})
	server.RegisterHandler(constants.ACCOUNT_UPGRADE, &upgradeHandler{})
	server.RegisterHandler(constants.ACCOUNT_PROFILE_USERINFO, &userinfoHandler{})
	server.RegisterHandler(constants.ACCOUNT_BIND_WX, &bindWXHandler{})
	server.RegisterHandler(constants.ACCOUNT_BIND_TG, &bindTGHandler{})

}
