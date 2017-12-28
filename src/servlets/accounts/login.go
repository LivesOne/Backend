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

type loginParam struct {
	Type    int    `json:"type"`
	UID     string `json:"uid"`
	EMail   string `json:"email"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	PWD     string `json:"pwd"`
	Key     string `json:"key"`
	Spkv    int    `json:"spkv"`
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
	UID    string            `json:"uid"`
	Token  string            `json:"token"`
	Expire int64             `json:"expire"`
	SPK    *responseLoginSPK `json:"spk"`
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
	aesKey, err := handler.parseAESKey(loginData.Param.Key)
	if err != nil || handler.isSignValid(aesKey, header.Signature, header.Timestamp) == false {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	var account *common.Account
	switch loginData.Param.Type {
	case constants.LOGIN_TYPE_UID:
		// right now, length of UID is 9
		if len(loginData.Param.UID) != constants.LEN_uid {
			logger.Info("login: uid info invalid")
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
		account, err = common.GetAccountByUID(loginData.Param.UID)
	case constants.LOGIN_TYPE_EMAIL:
		if utils.IsValidEmailAddr(loginData.Param.EMail) == false {
			response.SetResponseBase(constants.RC_EMAIL_NOT_MATCH)
			return
		}
		account, err = common.GetAccountByEmail(loginData.Param.EMail)
	case constants.LOGIN_TYPE_PHONE:
		account, err = common.GetAccountByPhone(loginData.Param.Country, loginData.Param.Phone)
	}

	if err != nil {
		logger.Info("login: read account from DB error:", err)
		response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
		return
	}
	logger.Info("read account form DB success:\n", utils.ToJSONIndent(account))

	if handler.checkUserPassword(account.LoginPassword, aesKey, loginData.Param.PWD, account.UIDString) == false {
		response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		return
	}

	// TODO:  get uid from the database
	// uid := strconv.FormatInt(account.UID, 10)
	const expire int64 = 24 * 3600
	newtoken, errNewT := token.New(account.UIDString, aesKey, expire)
	// newtoken, errNewT := token.New(handler.loginData.Param.UID, handler.aesKey, expire)
	if errNewT != constants.ERR_INT_OK {
		logger.Info("login: create token in cache error:", errNewT)
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	// newtoken, err = utils.RsaSign(newtoken, config.GetPrivateKeyFilename())
	iv := aesKey[:constants.AES_ivLen]
	key := aesKey[constants.AES_ivLen:]
	newtoken, err = utils.AesEncrypt(newtoken, string(key), string(iv))
	if err != nil {
		logger.Info("login: aes encrypt token error", err)
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	response.Data = &responseLogin{
		UID:    account.UIDString,
		Token:  newtoken,
		Expire: expire,
		// SPK: improve it later
		// SPK: &responseLoginSPK{
		// 	Ver:  1,
		// 	Key:  "",
		// 	Sign: "",
		// },
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

	if (loginData.Param.Type < constants.LOGIN_TYPE_UID) || (loginData.Param.Type > constants.LOGIN_TYPE_PHONE) {
		logger.Info("login: login type invalid")
		return false
	}

	if loginData.Param.Type == constants.LOGIN_TYPE_EMAIL && (utils.IsValidEmailAddr(loginData.Param.EMail) == false) {
		logger.Info("login: email info invalid")
		return false
	}

	if loginData.Param.Type == constants.LOGIN_TYPE_PHONE && (loginData.Param.Country == 0 || len(loginData.Param.Phone) < 1) {
		logger.Info("login: phone info invalid")
		return false
	}

	if (len(loginData.Param.PWD) < 1) || (len(loginData.Param.Key) < 1) {
		logger.Info("login: no pwd or key info")
		return false
	}

	if (len(loginData.Param.PWD) < 1) || (loginData.Param.Spkv < 1) {
		logger.Info("login: no pwd or spkv info")
		return false
	}

	return true
}

func (handler *loginHandler) isSignValid(aeskey, signature string, timestamp int64) bool {

	// signature := handler.header.Signature

	if len(signature) < 1 {
		return false
	}

	tmp := aeskey + strconv.FormatInt(timestamp, 10)
	hash := utils.Sha256(tmp)

	if signature == hash {
		logger.Info("login: verify header signature successful", signature, string(hash[:]))
	} else {
		logger.Info("login: verify header signature failed:", signature, string(hash[:]))
	}

	return signature == hash
}

// func (handler *loginHandler) verifySignature(signature, aeskey string, timestamp int64) bool {
// 	if len(signature) < 1 {
// 		return false
// 	}
// 	tmp := aeskey + strconv.FormatInt(timestamp, 10)
// 	// hash := sha256.Sum256([]byte(tmp))
// 	hash := utils.Sha256(tmp)
// 	return signature == string(hash[:])
// }

func (handler *loginHandler) parseAESKey(originalKey string) (string, error) {

	aeskey, err := utils.RsaDecrypt(originalKey, config.GetPrivateKey())
	if (err != nil) || (len(aeskey) != constants.AES_totalLen) {
		logger.Info("login: decrypt aes key error:", err)
		return "", err
	}

	logger.Info("login: aes key:", aeskey)

	return string(aeskey), nil
}

// func (handler *loginHandler) parsePWD(original string) (string, error) {

// 	aeskey, err := utils.RsaDecrypt(original, config.GetPrivateKey())
// 	if err != nil {
// 		logger.Info("login: decrypt pwd error:", err)
// 		return "", err
// 	}

// 	logger.Info("login: ----------hash pwd:", aeskey)

// 	return string(aeskey), nil
// }

func (handler *loginHandler) checkUserPassword(pwdInDB, aesKey, pwdUpload, uid string) bool {

	// const ivLen = 16
	// const keyLen = 32
	// 16 + 32 == 48
	if len(aesKey) != constants.AES_totalLen {
		logger.Info("login: invalid aes key", len(aesKey), aesKey)
		return false
	}

	iv := aesKey[:constants.AES_ivLen]
	key := aesKey[constants.AES_ivLen:]
	hashPwd, err := utils.AesDecrypt(pwdUpload, string(key), string(iv))
	if err != nil {
		logger.Info("login: invalid password")
		return false
	}

	pwd := utils.Sha256(hashPwd + uid)

	return (pwdInDB == pwd)
}
