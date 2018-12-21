package accounts

import (
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"golang.org/x/net/context"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
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
	if utils.SignValid(aesKey, header.Signature, header.Timestamp) == false {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	//uid := handler.getUID(aesKey, header.TokenHash, loginData.Param.Token)
	//// right now, length of UID is 9
	//if len(uid) != constants.LEN_uid {
	//	logger.Info("autologin: uid error")
	//	response.SetResponseBase(constants.RC_INVALID_TOKEN)
	//	return
	//}
	//
	//if !common.CheckUserLoginLimited(utils.Str2Int64(uid)) {
	//	response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
	//	return
	//}
	//
	//const expire int64 = 24 * 3600
	//errT := token.Update(header.TokenHash, aesKey, expire)
	//if errT != constants.ERR_INT_OK {
	//	logger.Info("autologin: update token hash failed")
	//	response.SetResponseBase(constants.RC_PARAM_ERR)
	//	return
	//}
	iv := aesKey[:constants.AES_ivLen]
	key := aesKey[constants.AES_ivLen:]

	tokenOriginal, err := utils.AesDecrypt(loginData.Param.Token, string(key), string(iv))

	if err != nil {
		logger.Info("logout: parse token failed:", err.Error())
		response.SetResponseBase(constants.RC_INVALID_TOKEN)
		return
	}

	uid, expire, code := autoLogin(header.TokenHash, tokenOriginal,loginData.Param.Key)
	if code == microuser.ResCode_OK {
		response.Data = &responseLogin{
			UID:    uid,
			Expire: expire,
		}
		rpc.ActiveUser(utils.Str2Int64(uid))
	} else {
		response.SetResponseBase(rpc.TokenErr2RcErr(code))
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

//func (handler *autoLoginHandler) getUID(aeskey, tokenHash, paramToken string) string {
//
//	// retrive the original token from cache
//	uid, _, tokenCache, errT := token.GetAll(tokenHash)
//	// logger.Info("autologin:====================", uid, tokenHash, paramToken)
//	if (errT != constants.ERR_INT_OK) || (len(uid) != constants.LEN_uid) {
//		logger.Info("autologin: get uid from token cache failed")
//		return ""
//	}
//
//	iv := aeskey[:constants.AES_ivLen]
//	key := aeskey[constants.AES_ivLen:]
//	tokenOriginal, err := utils.AesDecrypt(paramToken, string(key), string(iv))
//	if err != nil {
//		logger.Info("autologin: parse token failed", paramToken)
//		return ""
//	}
//
//	if tokenOriginal != tokenCache {
//		logger.Info("autologin: token invalid")
//		return ""
//	}
//
//	return uid
//}

func autoLogin(tokenHash, token, key string) (string, int64, microuser.ResCode) {
	cli := rpc.GetLoginClient()
	if cli != nil {
		req := &microuser.AutoLoginReq{
			TokenHash: tokenHash,
			Token:     token,
			Key:       key,
		}
		resp, err := cli.AutoLogin(context.Background(), req)
		if err != nil {
			logger.Error("grpc AutoLogin request error: ", err)
			return "", 0, microuser.ResCode_ERR_SYSTEM
		}
		if resp.Result != microuser.ResCode_OK {
			logger.Error("auto login failed ", resp.Result.String())
			return "", 0, resp.Result
		}
		return utils.Int642Str(resp.Uid), resp.Expire, resp.Result
	}
	return "", 0, microuser.ResCode_ERR_SYSTEM
}
