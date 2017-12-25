package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
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

	errT := token.Del(handler.header.TokenHash)
	if errT != constants.ERR_INT_OK {
		logger.Info("logout: delete token failed")
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

// func (handler *logoutHandler) getHashedToken(tokenUpload string) string {

// 	aesKey, errT := token.GetKey(handler.header.TokenHash)
// 	if errT != constants.ERR_INT_OK {
// 		logger.Info("autologin: get uid from token cache failed")
// 		return ""
// 	}

// 	iv := aesKey[:constants.AES_ivLen]
// 	key := aesKey[constants.AES_ivLen:]
// 	tokenTmp := utils.Base64Decode(tokenUpload)
// 	tokenDecrypt, err := utils.AesDecrypt(string(tokenTmp), string(key), string(iv))
// 	if err != nil {
// 		logger.Info("autologin: parse token failed", tokenUpload)
// 		return ""
// 	}

// 	return utils.Sha256(string(tokenDecrypt))
// }
