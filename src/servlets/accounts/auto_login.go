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
}

func (handler *autoLoginHandler) Method() string {
	return http.MethodPost
}

func (handler *autoLoginHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	loginData := new(autologinRequest)
	header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, loginData)

	if handler.checkRequestParams(header, loginData) == false {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	privKey, err := config.GetPrivateKey(loginData.Param.Spkv)
	if (err != nil) || (privKey == nil) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	aesKey, err := utils.RsaDecrypt(loginData.Param.Key, privKey)
	if (err != nil) || (len(aesKey) != constants.AES_totalLen) {
		logger.Info("autologin: decrypt aes key error:", err)
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}
	if handler.isSignValid(aesKey, header.Signature, header.Timestamp) == false {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := handler.getUID(aesKey, header.TokenHash, loginData.Param.Token)
	// right now, length of UID is 9
	if len(uid) != constants.LEN_uid {
		logger.Info("autologin: uid error")
		response.SetResponseBase(constants.RC_INVALID_TOKEN)
		return
	}

	if !common.CheckUserLoginLimited(utils.Str2Int64(uid)) {
		response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
		return
	}


	const expire int64 = 24 * 3600
	errT := token.Update(header.TokenHash, aesKey, expire)
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

func (handler *autoLoginHandler) checkRequestParams(header *common.HeaderParams, loginData *autologinRequest) bool {

	if (header == nil) || (header.IsValid() == false) {
		logger.Info("autologin: some header param missed")
		return false
	}

	if (loginData == nil) ||
		(loginData.Base.App == nil) ||
		(loginData.Base.App.IsValid() == false) {
		logger.Info("autologin: app info invalid")
		return false
	}

	if (len(loginData.Param.Token) < 1) ||
		(len(loginData.Param.Key) < 1) ||
		(loginData.Param.Spkv < 1) {
		logger.Info("autologin: no token or key or spkv info")
		return false
	}

	return true
}

func (handler *autoLoginHandler) isSignValid(aeskey, signature string, timestamp int64) bool {

	if len(signature) < 1 {
		logger.Info("autologin: no signature info")
		return false
	}

	tmp := aeskey + strconv.FormatInt(timestamp, 10)
	hash := utils.Sha256(tmp)

	if signature == hash {
		logger.Info("autologin: verify header signature successful", signature, string(hash[:]))
	} else {
		logger.Info("autologin: verify header signature failed:", signature, string(hash[:]))
	}

	return signature == hash
}

func (handler *autoLoginHandler) getUID(aeskey, tokenHash, paramToken string) string {

	// retrive the original token from cache
	uid, _, tokenCache, errT := token.GetAll(tokenHash)
	// logger.Info("autologin:====================", uid, tokenHash, paramToken)
	if (errT != constants.ERR_INT_OK) || (len(uid) != constants.LEN_uid) {
		logger.Info("autologin: get uid from token cache failed")
		return ""
	}

	iv := aeskey[:constants.AES_ivLen]
	key := aeskey[constants.AES_ivLen:]
	tokenOriginal, err := utils.AesDecrypt(paramToken, string(key), string(iv))
	if err != nil {
		logger.Info("autologin: parse token failed", paramToken)
		return ""
	}

	if tokenOriginal != tokenCache {
		logger.Info("autologin: token invalid")
		return ""
	}

	return uid
}
