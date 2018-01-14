package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type logoutParam struct {
	Token string `json:"token"`
}

type logoutRequest struct {
	Param logoutParam `json:"param"`
}

// logoutHandler implements the "Echo message" interface
type logoutHandler struct {
}

func (handler *logoutHandler) Method() string {
	return http.MethodPost
}

func (handler *logoutHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	logoutData := logoutRequest{}
	common.ParseHttpBodyParams(request, &logoutData)

	// if handler.checkRequestParams() == false {
	if (header.IsValid() == false) || (len(logoutData.Param.Token) < 1) {
		logger.Info("logout: header param missed or token info invalid")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if handler.checkToken(header.TokenHash, logoutData.Param.Token) == false {
		response.SetResponseBase(constants.RC_INVALID_TOKEN)
		return
	}

	errT := token.Del(header.TokenHash)
	if errT != constants.ERR_INT_OK {
		logger.Info("logout: remove token failed:", errT)
		response.SetResponseBase(constants.RC_INVALID_TOKEN)
	}
}

func (handler *logoutHandler) checkToken(headerTokenHash, paramToken string) bool {

	// retrive the original token from cache
	_, aesKey, tokenCache, errT := token.GetAll(headerTokenHash)
	if (errT != constants.ERR_INT_OK) || (len(aesKey) != constants.AES_totalLen) {
		logger.Info("logout: get token from cache failed: ", errT, len(aesKey))
		return false
	}

	iv := aesKey[:constants.AES_ivLen]
	key := aesKey[constants.AES_ivLen:]
	tokenOriginal, err := utils.AesDecrypt(paramToken, string(key), string(iv))
	// tokenTmp := utils.Base64Decode(tokenUpload)
	// tokenDecrypt, err := utils.AesDecrypt(string(tokenTmp), string(key), string(iv))
	if (err != nil) || (tokenOriginal != tokenCache) {
		logger.Info("logout: parse token failed:", err)
		return false
	}

	return true
}
