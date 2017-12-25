package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type logoutRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param string           `json:"param"`
}

// logoutHandler implements the "Echo message" interface
type logoutHandler struct {
	header     *common.HeaderParams // request header param
	logoutData *logoutRequest       // request login data
}

func (handler *logoutHandler) Method() string {
	return http.MethodPost
}

func (handler *logoutHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.logoutData)

	if handler.checkRequestParams() == false {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if handler.checkToken() == false {
		response.SetResponseBase(constants.RC_INVALID_TOKEN)
		return
	}

	errT := token.Del(handler.header.TokenHash)
	if errT != constants.ERR_INT_OK {
		logger.Info("logout: remove token failed")
		response.SetResponseBase(constants.RC_INVALID_TOKEN)
	}
}

func (handler *logoutHandler) checkRequestParams() bool {
	if (handler.header == nil) || (handler.logoutData == nil) || (handler.header.IsValid() == false) {
		return false
	}

	if (handler.logoutData.Base.App == nil) || (handler.logoutData.Base.App.IsValid() == false) {
		return false
	}

	if len(handler.logoutData.Param) < 1 {
		return false
	}

	return true
}

func (handler *logoutHandler) checkToken() bool {

	// retrive the original token from cache
	_, aesKey, tokenCache, errT := token.GetAll(handler.header.TokenHash)
	if (errT != constants.ERR_INT_OK) || (len(aesKey) != constants.AES_totalLen) {
		logger.Info("logout: get token from cache failed")
		return false
	}

	iv := aesKey[:constants.AES_ivLen]
	key := aesKey[constants.AES_ivLen:]
	tokenOriginal, err := utils.AesDecrypt(handler.logoutData.Param, string(key), string(iv))
	// tokenTmp := utils.Base64Decode(tokenUpload)
	// tokenDecrypt, err := utils.AesDecrypt(string(tokenTmp), string(key), string(iv))
	if (err != nil) || (tokenOriginal != tokenCache) {
		logger.Info("logout: parse token failed")
		return false
	}

	return true
}
