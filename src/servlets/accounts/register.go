package accounts

import (
	"server"
	"servlets/constants"
)

func RegisterHandlers() {
	server.RegisterHandler(constants.ACCOUNT_REGISTER, &registerUserHandler{})
	server.RegisterHandler(constants.ACCOUNT_LOGIN, &loginHandler{})
	server.RegisterHandler(constants.ACCOUNT_AUTOLOGIN, &autoLoginHandler{})
	server.RegisterHandler(constants.ACCOUNT_LOGOUT, &logoutHandler{})
}
