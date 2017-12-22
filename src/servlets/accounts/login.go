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
	sign string `json:"sign"`
}

type responseLogin struct {
	UID    string           `json:"uid"`
	Token  string           `json:"token"`
	Expire int64            `json:"expire"`
	SPK    responseLoginSPK `json:"spk"`
}

// loginHandler implements the "Echo message" interface
type loginHandler struct {
	header    *common.HeaderParams // request header param
	loginData *loginRequest        // request login data

	aesKey string // aes key (after parsing) uploaded by Client
	pwd    string // hashed user password
}

func (handler *loginHandler) Method() string {
	return http.MethodPost
}

func (handler *loginHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.loginData)

	if handler.checkRequestParams() == false {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	var err error
	handler.aesKey, err = handler.parseAESKey(handler.loginData.Param.Key)
	if err != nil || handler.isHeaderValid() == false {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	var account *common.Account
	switch handler.loginData.Param.Type {
	case constants.LOGIN_TYPE_UID:
		// right now, length of UID is 9
		if len(handler.loginData.Param.UID) != 9 {
			response.SetResponseBase(constants.RC_ACCOUNT_NOT_EXIST)
			return
		}
		account, err = common.GetAccountByUID(handler.loginData.Param.UID)
	case constants.LOGIN_TYPE_EMAIL:
	case constants.LOGIN_TYPE_PHONE:
	}

	if err != nil {
		response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
		return
	}
	logger.Info("read account form DB success:\n", utils.ToJSONIndent(account))

	if handler.checkUserPassword(account.LoginPassword, handler.aesKey, handler.loginData.Param.PWD) == false {
		response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		return
	}

	// TODO:  get uid from the database
	// uid := strconv.FormatInt(account.UID, 10)
	var expire int64 = 24 * 3600

	newtoken, errNewT := token.New(handler.loginData.Param.UID, handler.aesKey, expire)
	if errNewT != constants.ERR_INT_OK {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	newtoken, err = utils.RsaSign(newtoken, config.GetPrivateKeyFilename())
	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	response.Data = &responseLogin{
		UID:    handler.loginData.Param.UID,
		Token:  newtoken,
		Expire: expire,
	}
}

func (handler *loginHandler) checkRequestParams() bool {
	if handler.header == nil || (handler.loginData == nil) {
		return false
	}

	// const signLen = 64
	// const tokenHashLen = 64
	// if (handler.header.Timestamp < 1) || (len(handler.header.Signature) != signLen) || (len(handler.header.TokenHash) != tokenHashLen) {
	if handler.header.Timestamp < 1 || len(handler.header.Signature) < 1 || len(handler.header.TokenHash) < 1 {
		return false
	}

	if (handler.loginData.Base.App == nil) || (handler.loginData.Base.App.IsValid() == false) {
		return false
	}

	if (handler.loginData.Param.Type < constants.LOGIN_TYPE_UID) || (handler.loginData.Param.Type > constants.LOGIN_TYPE_PHONE) {
		return false
	}

	if handler.loginData.Param.Type == constants.LOGIN_TYPE_EMAIL && len(handler.loginData.Param.EMail) < 1 {
		return false
	}

	if handler.loginData.Param.Type == constants.LOGIN_TYPE_PHONE && (handler.loginData.Param.Country == 0 || len(handler.loginData.Param.Phone) < 1) {
		return false
	}

	if (len(handler.loginData.Param.PWD) < 1) || (len(handler.loginData.Param.Key) < 1) {
		return false
	}

	if (len(handler.loginData.Param.PWD) < 1) || (handler.loginData.Param.Spkv < 1) {
		return false
	}

	return true
}

func (handler *loginHandler) isHeaderValid() bool {

	signature := handler.header.Signature

	if len(signature) < 1 {
		return false
	}

	tmp := handler.aesKey + strconv.FormatInt(handler.header.Timestamp, 10)
	hash := utils.Sha256(tmp)

	if signature == string(hash[:]) {
		logger.Info("verify header signature successful", signature, string(hash[:]))
	} else {
		logger.Info("verify header signature failed:", signature, string(hash[:]))
	}

	return signature == string(hash[:])
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
	if err != nil {
		logger.Info("decrypt pwd error:", err)
		return "", err
	}

	logger.Info("----------aes key:", aeskey)

	return string(aeskey), nil
}

func (handler *loginHandler) parsePWD(original string) (string, error) {

	aeskey, err := utils.RsaDecrypt(original, config.GetPrivateKey())
	if err != nil {
		logger.Info("decrypt pwd error:", err)
		return "", err
	}

	logger.Info("----------hash pwd:", aeskey)

	return string(aeskey), nil
}

func (handler *loginHandler) checkUserPassword(pwdInDB, aesKey, pwdUpload string) bool {

	const ivLen = 16
	const keyLen = 32
	// len(iv) == 16
	// len(key) == 32
	// 16 + 32 == 48
	if len(aesKey) != (ivLen + keyLen) {
		logger.Info("invalide aes key")
		return false
	}

	iv := aesKey[:ivLen]
	key := aesKey[ivLen:]
	pwdUploadDecodeBase64 := utils.Base64Decode(pwdUpload)
	hashPwd, err := utils.AesEncrypt(string(pwdUploadDecodeBase64), string(key), string(iv))
	if err != nil {
		logger.Info("invalide password")
		return false
	}

	pwd := utils.Sha256(string(hashPwd) + handler.loginData.Param.UID)

	return (pwdInDB == pwd)
}
