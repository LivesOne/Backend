package accounts

import (
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"servlets/vcode"
	"utils"
	"utils/db_factory"
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
	common.ParseHttpBodyParams(request, requestData)

	if httpHeader.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if tokenErr != microuser.ResCode_OK {
		response.SetResponseBase(rpc.TokenErr2RcErr(tokenErr))
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}
	uid := utils.Str2Int64(uidString)

	// 解码 secret 参数
	secretString := requestData.Param.Secret
	secret := new(mailSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	if err := DecryptSecret(secretString, key, iv, secret); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}

	if !utils.IsValidEmailAddr(secret.Email) {
		response.SetResponseBase(constants.RC_EMAIL_NOT_MATCH)
		return
	}

	// 判断邮箱验证码正确
	ok, err := vcode.ValidateMailVCode(
		requestData.Param.VCodeId, requestData.Param.VCode, secret.Email)
	if ok == false {
		response.SetResponseBase(vcode.ConvImgErr(err))
		return
	}

	//if !common.CheckLoginPwd(uid, secret.Pwd) {
	//	response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
	//	return
	//}
	if f, _ := rpc.CheckPwd(uid, secret.Pwd, microuser.PwdCheckType_LOGIN_PWD); !f {
		response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		return
	}

	// save data to db
	f, dbErr := rpc.SetUserField(uid, microuser.UserField_EMAIL, secret.Email)
	if dbErr != nil || !f {
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
