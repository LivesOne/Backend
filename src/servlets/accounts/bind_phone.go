package accounts

import (
	"net/http"
	"encoding/json"

	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
)

type bindPhoneParam struct {
	VCodeId string `json:"vcode_id"`
	VCode  string `json:"vcode"`
	Secret string `json:"secret"`
}

type bindPhoneRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param bindPhoneParam `json:"param"`
}

// bindPhoneHandler
type bindPhoneHandler struct {
	header      *common.HeaderParams // request header param
	requestData *bindPhoneRequest    // request body
}

type phoneSecret struct {
	pwd string
	country int
	phone string
}

func (handler *bindPhoneHandler) Method() string {
	return http.MethodPost
}

func (handler *bindPhoneHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)
	requestData := new(bindPhoneRequest)
	common.ParseHttpBodyParams(request, &requestData)


	if httpHeader.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	tokenHash := httpHeader.TokenHash
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
	var secretData phoneSecret
	if err := json.Unmarshal([]byte(dataStr), &secretData); err != nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	country := secretData.country
	phone := secretData.phone

	// save data to db
	if err := common.SetPhone(uid, country, phone); err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
