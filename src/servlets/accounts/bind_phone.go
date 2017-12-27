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

type bindPhoneParam struct {
	VCodeId string `json:"vcode_id"`
	VCode   string `json:"vcode"`
	Secret  string `json:"secret"`
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
	Pwd     string
	Country int
	Phone   string
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
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
	}
	uid := utils.Str2Int64(uidString)

	// 解码 secret 参数
	secretString := requestData.Param.Secret
	secret := new(phoneSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	if err := DecryptSecret(secretString, key, iv, &secret); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}

	// 判断手机验证码正确
	ok, _ := vcode.ValidateSmsAndCallVCode(
		secret.Phone, secret.Country, requestData.Param.VCode, 0, 0)
	if ok == false {
		response.SetResponseBase(constants.RC_INVALID_VCODE)
		return
	}

	// save data to db
	dbErr := common.SetPhone(uid, secret.Country, secret.Phone)
	if dbErr != nil {
		if db_factory.CheckDuplicateByColumn(dbErr, "country") &&
			db_factory.CheckDuplicateByColumn(dbErr, "phone") {
			response.SetResponseBase(constants.RC_DUP_PHONE)
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
