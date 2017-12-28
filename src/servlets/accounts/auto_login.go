package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"strconv"
	"utils"
	"utils/config"
	"utils/logger"
)

type autologinParam struct {
	Token string `json:"token"`
	Key   string `json:"key"`
	Spkv  int    `json:"spkv"`
}

type autologinRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param autologinParam  `json:"param"`
}

type responseAutoLoginSPK struct {
	UID    string `json:"uid"`
	Expire int64  `json:"expire"`
}

type responseAutoLogin struct {
	UID    string                `json:"uid"`
	Expire int64                 `json:"expire"`
	SPK    *responseAutoLoginSPK `json:"spk"`
}

// autoLoginHandler implements the "Echo message" interface
type autoLoginHandler struct {
	header    *common.HeaderParams // request header param
	loginData *autologinRequest    // request login data

	aesKey string // aes key (after parsing) uploaded by Client
}

func (handler *autoLoginHandler) reset() {
	handler.header = nil
	handler.loginData = nil
	handler.aesKey = ""
}

func (handler *autoLoginHandler) Method() string {
	return http.MethodPost
}

func (handler *autoLoginHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	handler.reset()

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.loginData)

	if handler.checkRequestParams() == false {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	var err error
	handler.aesKey, err = utils.RsaDecrypt(handler.loginData.Param.Key, config.GetPrivateKey())
	if (err != nil) || (len(handler.aesKey) != constants.AES_totalLen) {
		logger.Info("autologin: decrypt aes key error:", err)
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}
	if handler.isSignValid() == false {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := handler.getUID()
	// right now, length of UID is 9
	if len(uid) != constants.LEN_uid {
		logger.Info("autologin: uid error")
		response.SetResponseBase(constants.RC_INVALID_TOKEN)
		return
	}

	const expire int64 = 24 * 3600
	errT := token.Update(handler.header.TokenHash, handler.aesKey, expire)
	if errT != constants.ERR_INT_OK {
		logger.Info("autologin: update token hash failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	response.Data = &responseLogin{
		UID:    uid,
		Expire: expire,
	}
}

func (handler *autoLoginHandler) checkRequestParams() bool {
	if (handler.header == nil) || (handler.loginData == nil) {
		return false
	}

	if handler.header.IsValid() == false {
		logger.Info("audologin: some header param missed")
		return false
	}

	if (handler.loginData.Base.App == nil) || (handler.loginData.Base.App.IsValid() == false) {
		logger.Info("autologin: app info invalid")
		return false
	}

	if (len(handler.loginData.Param.Token) < 1) ||
		(len(handler.loginData.Param.Key) < 1) ||
		(handler.loginData.Param.Spkv < 1) {
		logger.Info("augologin: no token or key or spkv info")
		return false
	}

	return true
}

func (handler *autoLoginHandler) isSignValid() bool {

	signature := handler.header.Signature

	if len(signature) < 1 {
		logger.Info("augologin: no signature info")
		return false
	}

	tmp := handler.aesKey + strconv.FormatInt(handler.header.Timestamp, 10)
	hash := utils.Sha256(tmp)

	if signature == hash {
		logger.Info("augologin: verify header signature successful", signature, string(hash[:]))
	} else {
		logger.Info("augologin: verify header signature failed:", signature, string(hash[:]))
	}

	return signature == hash
}

// func (handler *autoLoginHandler) getUID(tokenUpload string) (string, string) {

// 	iv := handler.aesKey[:constants.AES_ivLen]
// 	key := handler.aesKey[constants.AES_ivLen:]
// 	tokenTmp := utils.Base64Decode(tokenUpload)
// 	tokenDecrypt, err := utils.AesDecrypt(string(tokenTmp), string(key), string(iv))
// 	if err != nil {
// 		logger.Info("autologin: parse token failed", tokenUpload)
// 		return "", ""
// 	}

// 	uid, errT := token.GetUID(string(tokenDecrypt))
// 	if errT != constants.ERR_INT_OK {
// 		logger.Info("autologin: get uid from token cache failed", string(tokenDecrypt))
// 		return "", ""
// 	}

// 	return uid, utils.Sha256(string(tokenDecrypt))
// }

func (handler *autoLoginHandler) getUID() string {

	// retrive the original token from cache
	uid, _, tokenCache, errT := token.GetAll(handler.header.TokenHash)
	if (errT != constants.ERR_INT_OK) || (len(uid) != constants.LEN_uid) {
		logger.Info("autologin: get uid from token cache failed")
		return ""
	}

	iv := handler.aesKey[:constants.AES_ivLen]
	key := handler.aesKey[constants.AES_ivLen:]
	tokenOriginal, err := utils.AesDecrypt(handler.loginData.Param.Token, string(key), string(iv))
	if err != nil {
		logger.Info("autologin: parse token failed", handler.loginData.Param.Token)
		return ""
	}

	if tokenOriginal != tokenCache {
		logger.Info("autologin: token invalid")
		return ""
	}

	return uid
}
