package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/vcode"
)

type setTxPwdParam struct {
	Type   int    `json:"type"`
	VCodeId  string `json:"vcode_id"`
	VCode  string `json:"vcode"`
	PWD    string `json:"pwd"`
}

type setTxPwdRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param setTxPwdParam `json:"param"`
}

// setTxPwdHandler
type setTxPwdHandler struct {
	header      *common.HeaderParams // request header param
	requestData *setTxPwdRequest     // request body
}

func (handler *setTxPwdHandler) Method() string {
	return http.MethodPost
}

func (handler *setTxPwdHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)
	requestData := new(setTxPwdRequest)
	common.ParseHttpBodyParams(request, &requestData)

	if httpHeader.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, tokenErr := token.GetUID(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
	}
	uid := utils.Str2Int64(uidString)
	account, err := common.GetAccountByUID(uidString)
	if err != nil {
		response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
		return
	}

	// 检查验证码
	checkType := requestData.Param.Type
	if checkType == 1 {
		ok, err := vcode.ValidateMailVCode(
			requestData.Param.VCodeId, requestData.Param.VCode, account.Email)
		if ok == false {
			response.SetResponseBase(ValidateMailVCodeErr2RcErr(err))
			return
		}

	} else if checkType == 2 {
		ok, err := vcode.ValidateSmsAndCallVCode(
			account.Phone, account.Country, requestData.Param.VCode,0, 0)
		if err != nil || ok == false {
			response.SetResponseBase(constants.RC_INVALID_VCODE)
			return
		}

	} else {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// save to db
	if err := common.SetPaymentPassword(uid, requestData.Param.PWD); err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
