package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/config"
	"utils/db_factory"
	"utils/logger"
	"utils/vcode"
)

type bindPhoneParam struct {
	VCodeType int    `json:"vcode_type"`
	VCodeId   string `json:"vcode_id"`
	VCode     string `json:"vcode"`
	Secret    string `json:"secret"`
}

type bindPhoneRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param bindPhoneParam `json:"param"`
}

// bindPhoneHandler
type bindPhoneHandler struct {
}

type phoneSecret struct {
	Pwd     string
	Country int
	Phone   string
}

func (handler *bindPhoneHandler) Method() string {
	return http.MethodPost
}

func (handler *bindPhoneHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true,"bind phone")
	defer log.InfoAll()
	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	requestData := new(bindPhoneRequest)
	common.ParseHttpBodyParams(request, requestData)


	//校验参数合法
	if (header == nil) || !header.IsValid() ||
		(requestData == nil) ||
		(len(requestData.Param.Secret) < 1) ||
		(len(requestData.Param.VCodeId) < 1) ||
		(len(requestData.Param.VCode) < 1) {
		log.Error("bind phone: check param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}


	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(header.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		log.Error("bind phone: read user info error:", err)
		return
	}
	if !utils.SignValid(aesKey, header.Signature, header.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidString)

	if len(aesKey) != constants.AES_totalLen {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		log.Error("bind phone: read aes key from db error, length of aes key is:", len(aesKey))
		return
	}

	// 解码 secret 参数
	secretString := requestData.Param.Secret
	secret := new(phoneSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	if err := DecryptSecret(secretString, key, iv, secret); err != constants.RC_OK {
		response.SetResponseBase(err)
		log.Error("bind phone: Decrypt Secret error:", err)
		return
	}

	//如果这个参数为空，手动重置为下行短信
	vType:= requestData.Param.VCodeType
	if vType == 0 {
		vType = 1
	}


	switch vType {
	case 1:
		// 判断手机验证码正确
		if ok, c := vcode.ValidateSmsAndCallVCode(secret.Phone, secret.Country, requestData.Param.VCode, 0, vcode.FLAG_DEF);!ok {
			log.Error("bind phone: validate sms and call vcode failed")
			response.SetResponseBase(vcode.ConvSmsErr(c))
			return
		}
	case 2:
		if ok, resErr := vcode.ValidateSmsUpVCode(secret.Country, secret.Phone, requestData.Param.VCode); !ok {
			log.Info("validate up sms code failed")
			response.SetResponseBase(resErr)
			return
		}
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}



	account, err := common.GetAccountByUID(uidString)
	if err != nil {
		response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		return
	}
	// check login password
	pwd := utils.Sha256(secret.Pwd + uidString)
	if account.LoginPassword != pwd {
		response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		return
	}
	// check privilege
	limit := config.GetLimitByLevel(account.Level)
	if len(account.Phone) > 0 && !limit.ChangePhone(){
		response.SetResponseBase(constants.RC_USER_LEVEL_LIMIT)
		return
	}

	// save data to db
	dbErr := common.SetPhone(uid, secret.Country, secret.Phone)
	if dbErr != nil {
		// if db_factory.CheckDuplicateByColumn(dbErr, "country") &&
		// 	db_factory.CheckDuplicateByColumn(dbErr, "phone") {
		if db_factory.CheckDuplicateByColumn(dbErr, "mobile") {
			log.Error("bind phone: check phone duplicate error, dupped", dbErr)
			response.SetResponseBase(constants.RC_DUP_PHONE)
		} else {
			log.Error("bind phone: check phone duplicate error, other error", dbErr)
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		}
	}
}

