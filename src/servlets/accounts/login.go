package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"strconv"
	"utils"
	"utils/config"
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

type responseLogin struct {
	UID    string `json:"uid"`
	Token  string `json:"token"`
	Expire int64  `json:"expire"`
}

// loginHandler implements the "Echo message" interface
type loginHandler struct {
	header    *common.HeaderParams // request header param
	loginData *loginRequest        // request login data

	aesKey string // aes key uploaded by Client
}

func (handler *loginHandler) Method() string {
	return http.MethodPost
}

func (handler *loginHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.loginData)
	// if handler.readParams(request) == false {
	// 	// response.Base.RC = constants.RC_READ_REQUEST_PARAM_ERROR
	// 	response.Base.Msg = "read http request params error"
	// 	return
	// }

	handler.aesKey = handler.getAESKey(handler.loginData.Param.Key)
	if handler.verifySignature(handler.header.Signature, handler.aesKey, handler.header.Timestamp) == false {
		// response.Base.RC = constants.RC_INVALID_SIGNATURE
		response.Base.Msg = "invalid signature"
		return
	}

	switch handler.loginData.Param.Type {
	case constants.LOGIN_TYPE_UID:
	case constants.LOGIN_TYPE_EMAIL:
	case constants.LOGIN_TYPE_PHONE:
	}

	// TODO:  get uid from the database
	uid := "123456789"
	var expire int64 = 24 * 3600

	newtoken, errNewT := token.New(uid, handler.aesKey, expire)
	if errNewT != constants.ERR_INT_OK {
		response.Base.RC = constants.RC_SYSTEM_ERR.Rc
		response.Base.Msg = constants.RC_SYSTEM_ERR.Msg
		return
	}

	newtoken, err := utils.RsaSign(newtoken, config.GetConfig().PrivKey)
	if err != nil {
		response.Base.RC = constants.RC_SYSTEM_ERR.Rc
		response.Base.Msg = constants.RC_SYSTEM_ERR.Msg
		return
	}

	response.Data = &responseLogin{
		UID:    uid,
		Token:  newtoken,
		Expire: expire,
	}
}

// func (handler *loginHandler) readParams(request *http.Request) bool {

// 	handler.header = common.ParseHttpHeaderParams(request)
// 	if (handler.header.Timestamp < 0) || (len(handler.header.Signature) < 1) {
// 		return false
// 	}

// 	common.ParseHttpBodyParams(request, &handler.loginData)

// 	return true
// }

func (handler *loginHandler) verifySignature(signature, aeskey string, timestamp int64) bool {
	if len(signature) < 1 {
		return false
	}
	tmp := aeskey + strconv.FormatInt(timestamp, 10)
	// hash := sha256.Sum256([]byte(tmp))
	hash := utils.Sha256(tmp)
	return signature == string(hash[:])
}

func (handler *loginHandler) getAESKey(originalKey string) string {

	// decodedBase64, err := base64.StdEncoding.DecodeString(originalKey)
	// if err != nil {
	// 	// logger.Info("decode key error:", err, originalKey)
	// 	logger.Info("decode key error, base64:", err)
	// 	return ""
	// }

	aeskey, _ := utils.RsaDecrypt(originalKey, config.GetPrivateKey())
	// if err != nil {
	// 	// logger.Info("decode key error:", err, originalKey)
	// 	logger.Info("decode key error, rsa:", err)
	// 	return ""
	// }

	return string(aeskey)
}
