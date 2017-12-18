package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strconv"
	"time"
	"utils"
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
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(handler.response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.registerData)

	if handler.checkRequestParams() == false {
		handler.setResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// fmt.Println("registerUserHandler) Handle", msg)
	// hashPwd := utils.RsaDecrypt(handler.registerData.Param.PWD, config.GetConfig().PrivKey)

	account := handler.getAccount()

	var err error

	switch handler.registerData.Param.Type {
	case constants.LOGIN_TYPE_UID:
		_, err = common.InsertAccount(account)
	case constants.LOGIN_TYPE_EMAIL:
		if common.ExistsEmail(account.Email) {
			handler.setResponseBase(constants.RC_DUP_EMAIL)
			return
		} else {
			_, err = common.InsertAccountWithEmail(account)
		}
	case constants.LOGIN_TYPE_PHONE:
		if common.ExistsPhone(account.Country, account.Phone) {
			handler.setResponseBase(constants.RC_DUP_PHONE)
			return
		} else {
			_, err = common.InsertAccountWithPhone(account)
		}
	}

	if err != nil {
		handler.setResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	handler.response.Data = &responseRegister{
		UID:     account.UIDString,
		Regtime: account.RegisterTime,
	}
}

func (handler *registerUserHandler) setResponseBase(error constants.Error) {
	handler.response.Base.RC = error.Rc
	handler.response.Base.Msg = error.Msg
	logger.Info(error.Msg)
}

func (handler *registerUserHandler) checkRequestParams() bool {
	if handler.header.Timestamp < 1 {
		return false
	}

	if (handler.registerData.Base.App == nil) || (handler.registerData.Base.App.IsValid() == false) {
		return false
	}

	if (handler.registerData.Param.Type < constants.LOGIN_TYPE_UID) || (handler.registerData.Param.Type > constants.LOGIN_TYPE_PHONE) {
		return false
	}

	if handler.registerData.Param.Type == constants.LOGIN_TYPE_EMAIL && len(handler.registerData.Param.EMail) < 1 {
		return false
	}

	if handler.registerData.Param.Type == constants.LOGIN_TYPE_PHONE && (handler.registerData.Param.Country == 0 || len(handler.registerData.Param.Phone) < 1) {
		return  false
	}

	if (len(handler.registerData.Param.PWD) < 1) || (handler.registerData.Param.Spkv < 1) {
		return false
	}

	return true
}

func (handler *registerUserHandler) getAccount() common.Account {
	var account common.Account
	var uid string
	var uid_num int64

	for {
		uid = common.GenerateUID(9)
		uid_num, _ = strconv.ParseInt(uid, 10, 64)

		if common.ExistsUID(uid_num) {
			continue
		} else {
			break
		}
	}

	account.UIDString = uid
	account.UID = uid_num

	account.Email = handler.registerData.Param.EMail
	account.Country = handler.registerData.Param.Country
	account.Phone = handler.registerData.Param.Phone

	account.LoginPassword = utils.Sha256(handler.registerData.Param.PWD + uid)
	account.RegisterTime = time.Now().Unix()
	account.UpdateTime = account.RegisterTime
	account.RegisterType = handler.registerData.Param.Type

	return account
}
