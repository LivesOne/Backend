package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/vcode"
	"utils/logger"
)

type setTxPwdParam struct {
	Type    int    `json:"type"`
	VCodeId string `json:"vcode_id"`
	VCode   string `json:"vcode"`
	PWD     string `json:"pwd"`
}

type setTxPwdRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param setTxPwdParam `json:"param"`
}

// setTxPwdHandler
type setTxPwdHandler struct {
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
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
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
			response.SetResponseBase(vcode.ConvImgErr(err))
			return
		}

	} else if checkType == 2 {
		ok, err := vcode.ValidateSmsAndCallVCode(
			account.Phone, account.Country, requestData.Param.VCode, 0, 0)
		if  ok == false {
			response.SetResponseBase(vcode.ConvSmsErr(err))
			return
		}

	} else {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 解析出“sha256(密码)”
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	pwdSha256, err := utils.AesDecrypt(requestData.Param.PWD, key, iv)
	logger.Debug("pwdSha256",pwdSha256)
	if err != nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 数据库实际保存的密码格式为“sha256(sha256(密码) + uid)”
	pwdDb := utils.Sha256(pwdSha256 + uidString)

	// save to db
	if err := common.SetPaymentPassword(uid, pwdDb); err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
