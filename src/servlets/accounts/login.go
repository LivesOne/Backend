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

type loginParam struct {
	// Type    int    `json:"type"`
	// UID     string `json:"uid"`
	// EMail   string `json:"email"`
	// Country int    `json:"country"`
	// Phone   string `json:"phone"`
	PWD     string `json:"pwd"`
	Key     string `json:"key"`
	Spkv    int    `json:"spkv"`
	Account string `json:"account"`
}

type loginRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param loginParam      `json:"param"`
}

type responseLoginSPK struct {
	Ver  int    `json:"ver"`
	Key  string `json:"key"`
	Sign string `json:"sign"`
}

type responseLogin struct {
	UID    string `json:"uid"`
	Token  string `json:"token,omitempty"`
	Expire int64  `json:"expire"`
	Cookie string `json:"cookie"`
	//SPK    *responseLoginSPK `json:"spk"`
}

type limitedRes struct {
	Uid string `json:"uid"`
}

type tmpLimitedRes struct {
	LimitTime int `json:"limit_time"`
}

// loginHandler implements the "Echo message" interface
type loginHandler struct {
}

func (handler *loginHandler) Method() string {
	return http.MethodPost
}

func (handler *loginHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	loginData := loginRequest{}
	common.ParseHttpBodyParams(request, &loginData)

	if handler.checkRequestParams(header, &loginData) == false {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// var err error
	aesKey, err := handler.parseAESKey(loginData.Param.Key, loginData.Param.Spkv)
	if err != nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	if utils.SignValid(aesKey, header.Signature, header.Timestamp) == false {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	cli := rpc.GetLoginClient()
	if cli == nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	privKey, err := config.GetPrivateKey(loginData.Param.Spkv)
	if (err != nil) || (privKey == nil) {
		logger.Info("login: get private key by spkv err:", err)
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	aeskey, err := utils.RsaDecrypt(loginData.Param.Key, privKey)
	if (err != nil) || (len(aeskey) != constants.AES_totalLen) {
		logger.Info("login: decrypt aes key error:", err)
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	iv := aeskey[:constants.AES_ivLen]
	key := aesKey[constants.AES_ivLen:]
	logger.Info("login aeskey", aesKey)
	hashPwd, err := utils.AesDecrypt(loginData.Param.PWD, string(key), string(iv))
	if err != nil {
		logger.Info("login: invalid password")
		response.SetResponseBase(constants.RC_PARAM_ERR)
	}

	req := &microuser.LoginUserReq{
		Account: loginData.Param.Account,
		PwdHash: hashPwd,
		Key:     aesKey,
		App:     transBasAppInfo(loginData.Base.App.AppID),
		Plat:    utils.Int2Str(loginData.Base.App.Plat),
	}

	resp, err := cli.Login(context.Background(), req)
	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if resp.Result != microuser.ResCode_OK {
		switch resp.Result {
		case microuser.ResCode_ERR_LIMITED:
			response.SetResponseBase(constants.RC_ACCOUNT_LIMITED)
			response.Data = limitedRes{
				Uid: resp.Uid,
			}
			return
		case microuser.ResCode_ERR_NOTFOUND:
			response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
			return
		default:
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
	}
	newtoken, err := utils.AesEncrypt(resp.Token, string(key), string(iv))
	if err != nil {
		logger.Info("login: aes encrypt token error", err)
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	rpc.ActiveUser(utils.Str2Int64(resp.Uid))
	cookie, _ := common.GetCookieByTokenAndKey(newtoken, aeskey,resp.Uid)
	response.Data = &responseLogin{
		UID:    resp.Uid,
		Token:  newtoken,
		Expire: resp.Expire,
		Cookie: cookie,
	}
}

func (handler *loginHandler) checkRequestParams(header *common.HeaderParams, loginData *loginRequest) bool {
	if header == nil || (loginData == nil) {
		return false
	}

	// const signLen = 64
	// const tokenHashLen = 64
	if (header.IsValidTimestamp() == false) || (header.IsValidSign() == false) {
		logger.Info("login: some header param missed")
		return false
	}

	if (loginData.Base.App == nil) || (loginData.Base.App.IsValid() == false) {
		logger.Info("login: app info invalid")
		return false
	}

	if len(loginData.Param.Account) < 1 {
		logger.Info("login: account param missed")
		return false
	}

	if (len(loginData.Param.PWD) < 1) || (len(loginData.Param.Key) < 1) {
		logger.Info("login: no pwd or key info")
		return false
	}

	if loginData.Param.Spkv < 1 {
		logger.Info("login: no pwd or spkv info")
		return false
	}

	return true
}

func (handler *loginHandler) parseAESKey(originalKey string, spkv int) (string, error) {

	privKey, err := config.GetPrivateKey(spkv)
	if (err != nil) || (privKey == nil) {
		return "", err
	}
	aeskey, err := utils.RsaDecrypt(originalKey, privKey)
	if (err != nil) || (len(aeskey) != constants.AES_totalLen) {
		logger.Info("login: decrypt aes key error:", err)
		return "", err
	}

	logger.Info("login: aes key:", aeskey)

	return string(aeskey), nil
}

func transBasAppInfo(app interface{}) string {
	appid := ""
	switch app.(type) {
	case int:
		appid = utils.Int2Str(app.(int))
	case int64:
		appid = utils.Int642Str(app.(int64))
	case float64:
		appid = utils.Float642Str(app.(float64))
	case string:
		appid = app.(string)
	}
	return appid
}
