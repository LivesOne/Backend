package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/db_factory"
	"utils/vcode"
)

type bindEMailParam struct {
	VCodeId string `json:"vcode_id"`
	VCode   string `json:"vcode"`
	Secret  string `json:"secret"`
}

type bindEMailRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param bindEMailParam `json:"param"`
}

// bindEMailHandler
type bindEMailHandler struct {
	header      *common.HeaderParams // request header param
	requestData *bindEMailRequest    // request body
}

type mailSecret struct {
	Pwd   string
	Email string
}

func (handler *bindEMailHandler) Method() string {
	return http.MethodPost
}

func (handler *bindEMailHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)
	requestData := new(bindEMailRequest)
	common.ParseHttpBodyParams(request, &requestData)

	if httpHeader.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}
	uid := utils.Str2Int64(uidString)

	// 解码 secret 参数
	secretString := requestData.Param.Secret
	secret := new(mailSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	if err := DecryptSecret(secretString, key, iv, &secret); err != constants.RC_OK {
		response.SetResponseBase(err)
	}

	if utils.IsValidEmailAddr(secret.Email) {
		response.SetResponseBase(constants.RC_INVALIDE_EMAIL_ADDRESS)
		return
	}

	// 判断邮箱验证码正确
	ok, err := vcode.ValidateMailVCode(
		requestData.Param.VCodeId, requestData.Param.VCode, secret.Email)
	if ok == false {
		response.SetResponseBase(ValidateMailVCodeErr2RcErr(err))
		return
	}

	// save data to db
	dbErr := common.SetEmail(uid, secret.Email)
	if dbErr != nil {
		if db_factory.CheckDuplicateByColumn(dbErr, "email") {
			response.SetResponseBase(constants.RC_DUP_EMAIL)
			return
		} else {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
