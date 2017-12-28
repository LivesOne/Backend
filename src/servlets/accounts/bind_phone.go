package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/db_factory"
	"utils/logger"
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
	// header      *common.HeaderParams // request header param
	// requestData *bindPhoneRequest // request body
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

	header := common.ParseHttpHeaderParams(request)
	requestData := new(bindPhoneRequest)
	common.ParseHttpBodyParams(request, requestData)

	if handler.checkRequestParams(header, requestData) == false {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(header.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		logger.Info("bind phone: read user info error:", err)
		return
	}
	uid := utils.Str2Int64(uidString)

	if len(aesKey) != constants.AES_totalLen {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		logger.Info("bind phone: read aes key from db error, length of aes key is:", len(aesKey))
		return
	}

	// 解码 secret 参数
	secretString := requestData.Param.Secret
	secret := new(phoneSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	if err := DecryptSecret(secretString, key, iv, &secret); err != constants.RC_OK {
		response.SetResponseBase(err)
		logger.Info("bind phone: Decrypt Secret error:", err)
		return
	}

	// 判断手机验证码正确
	ok, _ := vcode.ValidateSmsAndCallVCode(
		secret.Phone, secret.Country, requestData.Param.VCode, 0, 0)
	if ok == false {
		logger.Info("bind phone: validate sms and call vcode failed")
		response.SetResponseBase(constants.RC_DUP_PHONE)
		return
	}

	// save data to db
	dbErr := common.SetPhone(uid, secret.Country, secret.Phone)
	if dbErr != nil {
		if db_factory.CheckDuplicateByColumn(dbErr, "country") &&
			db_factory.CheckDuplicateByColumn(dbErr, "phone") {
			response.SetResponseBase(constants.RC_DUP_PHONE)
		} else {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		}
	}
}

func (handler *bindPhoneHandler) checkRequestParams(header *common.HeaderParams, requestData *bindPhoneRequest) bool {

	if (header == nil) || (header.IsValid() == false) {
		logger.Info("bind phone: invalid header info")
		return false
	}

	if (requestData == nil) ||
		(len(requestData.Param.Secret) < 1) ||
		(len(requestData.Param.VCodeId) < 1) ||
		(len(requestData.Param.VCode) < 1) {
		logger.Info("bind phone: no enough paramter")
		return false
	}

	return true
}
