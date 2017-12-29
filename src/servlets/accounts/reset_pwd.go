package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
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

	if header.IsValidTimestamp() == false {
		logger.Info("reset password: invalid timestamp")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	// uidString, _, _, tokenErr := token.GetAll(header.TokenHash)
	// if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
	// 	response.SetResponseBase(err)
	// }
	// uid := utils.Str2Int64(uidString)
	// account, err := common.GetAccountByUID(uidString)
	// if err != nil {
	// 	response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
	// 	return
	// }

	var account *common.Account
	var err error
	// 检查验证码
	checkType := requestData.Param.Type
	if checkType == 1 {
		if utils.IsValidEmailAddr(requestData.Param.EMail) == false {
			logger.Info("reset password: invalid email address")
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		account, err = common.GetAccountByEmail(requestData.Param.EMail)
		if (err != nil) || (account == nil) {
			logger.Info("reset password: get account info by email failed:", requestData.Param.EMail)
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
		ok, errT := vcode.ValidateMailVCode(
			requestData.Param.VCodeID, requestData.Param.VCode, account.Email)
		if ok == false {
			logger.Info("reset password: verify email vcode failed", errT)
			response.SetResponseBase(ValidateMailVCodeErr2RcErr(errT))
			return
		}

	} else if checkType == 2 {
		if (len(requestData.Param.Phone) < 1) || (requestData.Param.Country < 1) {
			logger.Info("reset password: invalid phone or country", requestData.Param.Country, requestData.Param.Phone)
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		account, err = common.GetAccountByPhone(requestData.Param.Country, requestData.Param.Phone)
		if (err != nil) || (account == nil) {
			logger.Info("reset password: get account info by phone failed:", requestData.Param.Country, requestData.Param.Phone)
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
		ok, err := vcode.ValidateSmsAndCallVCode(
			account.Phone, account.Country, requestData.Param.VCode, 0, 0)
		if err != nil || ok == false {
			logger.Info("reset password: verify sms vcode failed", err)
			response.SetResponseBase(constants.RC_INVALID_VCODE)
			return
		}

	} else {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 解析出“sha256(密码)”
	privKey, err := config.GetPrivateKey(requestData.Param.Spkv)
	if (err != nil) || (privKey == nil) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	pwdSha256, err := utils.RsaDecrypt(requestData.Param.PWD, privKey)
	if err != nil {
		response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		return
	}

	// 数据库实际保存的密码格式为“sha256(sha256(密码) + uid)”
	pwdDb := utils.Sha256(pwdSha256 + account.UIDString)

	// save to db
	if err := common.SetLoginPassword(account.UID, pwdDb); err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
