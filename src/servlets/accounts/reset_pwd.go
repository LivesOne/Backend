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
	log := logger.NewLvtLogger(true, "reset login pwd")
	defer log.InfoAll()
	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	requestData := resetPwdRequest{}
	common.ParseHttpBodyParams(request, &requestData)

	if header.IsValidTimestamp() == false {
		log.Info("reset password: invalid timestamp")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	var account *common.Account
	var err error
	// 检查验证码
	checkType := requestData.Param.Type
	switch checkType {
	case 1: //邮箱验证
		if !utils.IsValidEmailAddr(requestData.Param.EMail) {
			log.Info("reset password: invalid email address")
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}

		if ok, errT := vcode.ValidateMailVCode(requestData.Param.VCodeID, requestData.Param.VCode, requestData.Param.EMail); !ok {
			log.Info("reset password: verify email vcode failed", errT)
			response.SetResponseBase(vcode.ConvImgErr(errT))
			return
		}

		account, err = common.GetAccountByEmail(requestData.Param.EMail)
		if (err != nil) || (account == nil) {
			log.Info("reset password: get account info by email failed:", requestData.Param.EMail)
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
	case 2: //短信下行验证
		if (len(requestData.Param.Phone) < 1) || (requestData.Param.Country < 1) {
			log.Info("reset password: invalid phone or country", requestData.Param.Country, requestData.Param.Phone)
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		if ok, err := vcode.ValidateSmsAndCallVCode(requestData.Param.Phone, requestData.Param.Country, requestData.Param.VCode, 0, 0); !ok {
			e := vcode.ConvSmsErr(err)
			log.Info("reset password: verify sms vcode failed", e)
			response.SetResponseBase(e)
			return
		}
		account, err = common.GetAccountByPhone(requestData.Param.Country, requestData.Param.Phone)
		if (err != nil) || (account == nil) {
			log.Info("reset password: get account info by phone failed:", requestData.Param.Country, requestData.Param.Phone)
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
	case 3: //短信上行验证
		if (len(requestData.Param.Phone) < 1) || (requestData.Param.Country < 1) {
			log.Info("reset password: invalid phone or country", requestData.Param.Country, requestData.Param.Phone)
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		if ok, resErr := vcode.ValidateSmsUpVCode(requestData.Param.Country, requestData.Param.Phone, requestData.Param.VCode); !ok {
			log.Info("validate up sms code failed")
			response.SetResponseBase(resErr)
			return
		}
		account, err = common.GetAccountByPhone(requestData.Param.Country, requestData.Param.Phone)
		if (err != nil) || (account == nil) {
			log.Info("reset password: get account info by phone failed:", requestData.Param.Country, requestData.Param.Phone)
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return

	}

	// 解析出“sha256(密码)”
	privKey, err := config.GetPrivateKey(requestData.Param.Spkv)
	if (err != nil) || (privKey == nil) {
		log.Error("can not get private key")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	pwdSha256, err := utils.RsaDecrypt(requestData.Param.PWD, privKey)
	if err != nil {
		log.Info("reset password: decrypt pwd error:", err)
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	// 数据库实际保存的密码格式为“sha256(sha256(密码) + uid)”
	pwdDb := utils.Sha256(pwdSha256 + account.UIDString)

	// save to db
	if err := common.SetLoginPassword(account.UID, pwdDb); err != nil {
		log.Info("reset password: save login pwd in DB error:", err)
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
}
