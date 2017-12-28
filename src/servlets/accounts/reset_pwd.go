package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/config"
	"utils/logger"
	"utils/vcode"
)

type resetPwdParam struct {
	Type    int    `json:"type"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	VCodeID string `json:"vcode_id"`
	VCode   string `json:"vcode"`
	PWD     string `json:"pwd"`
	Spkv    int    `json:"spkv"`
}

type resetPwdRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param resetPwdParam   `json:"param"`
}

// resetPwdHandler
type resetPwdHandler struct {
}

func (handler *resetPwdHandler) Method() string {
	return http.MethodPost
}

func (handler *resetPwdHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	requestData := resetPwdRequest{}
	common.ParseHttpBodyParams(request, &requestData)

	if (header.IsValidTimestamp() == false) || (header.IsValidTokenhash() == false) {
		logger.Info("reset password: some header param missed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, _, _, tokenErr := token.GetAll(header.TokenHash)
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
		if (utils.IsValidEmailAddr(requestData.Param.EMail) == false) ||
			(requestData.Param.EMail != account.Email) {
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		ok, err := vcode.ValidateMailVCode(
			requestData.Param.VCodeID, requestData.Param.VCode, account.Email)
		if ok == false {
			response.SetResponseBase(ValidateMailVCodeErr2RcErr(err))
			return
		}

	} else if checkType == 2 {
		if (len(requestData.Param.Phone) < 1) ||
			(requestData.Param.Country < 1) ||
			(requestData.Param.Country != account.Country) ||
			(requestData.Param.Phone != account.Phone) {
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		ok, err := vcode.ValidateSmsAndCallVCode(
			account.Phone, account.Country, requestData.Param.VCode, 0, 0)
		if err != nil || ok == false {
			response.SetResponseBase(constants.RC_INVALID_VCODE)
			return
		}

	} else {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 解析出“sha256(密码)”
	pwdSha256, err := utils.RsaDecrypt(requestData.Param.PWD, config.GetPrivateKey())
	if err != nil {
		response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		return
	}

	// 数据库实际保存的密码格式为“sha256(sha256(密码) + uid)”
	pwdDb := utils.Sha256(pwdSha256 + uidString)

	// save to db
	if err := common.SetLoginPassword(uid, pwdDb); err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}

// func (handler *resetPwdHandler) checkRequestParams() bool {

// 	if (header.IsValidTimestamp() == false) || (header.IsValidTokenhash() == false) {
// 		logger.Info("reset password: some header param missed")
// 		return false
// 	}

// 	return true
// }
