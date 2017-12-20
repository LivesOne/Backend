package accounts

import (
	"net/http"
	"encoding/json"

	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
)

type bindEMailParam struct {
	VCodeId string `json:"vcode_id"`
	VCode  string `json:"vcode"`
	Secret string `json:"secret"`
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
	pwd string
	email string
}

func (handler *bindEMailHandler) Method() string {
	return http.MethodPost
}

func (handler *bindEMailHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.requestData)

	if handler.header.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	tokenHash := handler.header.TokenHash
	uidString, tokenErr := token.GetUID(tokenHash)
	switch tokenErr {
		case constants.ERR_INT_OK:
			break
		case constants.ERR_INT_TK_DB:
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		case constants.ERR_INT_TK_DUPLICATE:
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		case constants.ERR_INT_TK_NOTEXISTS:
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		default:
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
	}
	uid := utils.Str2Int64(uidString)

	// 判断验证码正确

	// 解码 secret 参数
	key := ""
	iv := ""
	secret := handler.requestData.Param.Secret
	dataStr, err := utils.AesDecrypt(secret, key, iv)
	if err != nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	var secretData mailSecret
	if err := json.Unmarshal([]byte(dataStr), &secretData); err != nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	email := secretData.email

	// save data to db
	if err := common.SetEmail(uid, email); err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
