package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strconv"
	"time"
	"utils/logger"
)

// registerParam holds the request "param" field
type registerParam struct {
	Type    int    `json:"type"`
	Action  string `json:"action"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	VCode   string `json:"vcode"`
	PWD     string `json:"pwd"`
	Spkv    int    `json:"spkv"`
}

// registerRequest holds entire request data
type registerRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param registerParam   `json:"param"`
}

// responseData holds response "data" field
type responseRegister struct {
	UID     string `json:"uid"`
	Regtime int64  `json:"regtime"`
}

// registerUserHandler implements the "Echo message" interface
type registerUserHandler struct {
	// http request, header params
	header *common.HeaderParams
	// http request, body params
	registerData *registerRequest

	// http response data to client
	response *common.ResponseData
}

func (handler *registerUserHandler) Method() string {
	return http.MethodPost
}

func (handler *registerUserHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	handler.response = &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK,
			Msg: "ok",
		},
	}
	defer common.FlushJSONData2Client(handler.response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.registerData)
	if handler.checkRequestParams() == false {
		return
	}

	// fmt.Println("registerUserHandler) Handle", msg)
	// hashPwd := utils.RsaDecrypt(handler.registerData.Param.PWD, config.GetConfig().PrivKey)

	account := handler.getAccount()

	// newtoken, err := token.New(uid, "key", 24*3600)

	switch handler.registerData.Param.Type {
	case constants.LOGIN_TYPE_UID:
		common.InsertAccount(account)
	case constants.LOGIN_TYPE_EMAIL:
	case constants.LOGIN_TYPE_PHONE:
	}

	handler.response.Data = &responseRegister{
		UID:     account.UIDString,
		Regtime: account.RegisterTime,
	}
}

func (handler *registerUserHandler) checkRequestParams() bool {
	// if handler.header.Timestamp < 1 {
	// 	handler.response.Base.RC = constants.RC_READ_REQUEST_PARAM_ERROR
	// 	handler.response.Base.Msg = "read http request params error"
	// 	return false
	// }

	// if (handler.registerData.Base.App == nil) || (handler.registerData.Base.App.IsValid() == false) {
	// 	handler.response.Base.RC = constants.RC_READ_REQUEST_PARAM_ERROR
	// 	handler.response.Base.Msg = "read http request params error"
	// 	return false
	// }

	// if (handler.registerData.Param.Type < constants.LOGIN_TYPE_UID) && (handler.registerData.Param.Type > constants.LOGIN_TYPE_PHONE) {
	// 	handler.response.Base.RC = constants.RC_READ_REQUEST_PARAM_ERROR
	// 	handler.response.Base.Msg = "read http request params error"
	// 	return false
	// }

	// if (len(handler.registerData.Param.PWD) < 1) && (handler.registerData.Param.Spkv < 1) {
	// 	handler.response.Base.RC = constants.RC_READ_REQUEST_PARAM_ERROR
	// 	handler.response.Base.Msg = "read http request params error"
	// 	return false
	// }

	if (handler.header.Timestamp < 1) ||
		(handler.registerData.Base.App == nil) || (handler.registerData.Base.App.IsValid() == false) ||
		(handler.registerData.Param.Type < constants.LOGIN_TYPE_UID) && (handler.registerData.Param.Type > constants.LOGIN_TYPE_PHONE) ||
		(len(handler.registerData.Param.PWD) < 1) && (handler.registerData.Param.Spkv < 1) {

		handler.response.Base.RC = constants.RC_READ_REQUEST_PARAM_ERROR
		handler.response.Base.Msg = "request params error"
		logger.Info("request params error")
		return false
	}

	return true
}

func (handler *registerUserHandler) getAccount() common.Account {
	var account common.Account

	uid := common.GenerateUID(9)

	account.UIDString = uid
	account.UID, _ = strconv.ParseInt(uid, 10, 64)

	account.Email = handler.registerData.Param.EMail
	account.Country = handler.registerData.Param.Country
	account.Phone = handler.registerData.Param.Phone

	account.LoginPassword = handler.registerData.Param.PWD
	account.RegisterTime = time.Now().Unix()
	account.UpdateTime = account.RegisterTime
	account.RegisterType = handler.registerData.Param.Type

	return account
}
