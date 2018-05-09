package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

const (
	LOGIN_PASSWORD   = 1
	PAYMENT_PASSWORD = 2
)

type modifyPwdParam struct {
	Type   int    `json:"type"`
	Secret string `json:"secret"`
}

type modifySecret struct {
	Pwd    string `json:"pwd"`
	NewPwd string `json:"new_pwd"`
}

type modifyPwdRequest struct {
	Param modifyPwdParam `json:"param"`
}

// modifyPwdHandler
type modifyPwdHandler struct {
}

func (handler *modifyPwdHandler) Method() string {
	return http.MethodPost
}

func (handler *modifyPwdHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)
	requestData := new(modifyPwdRequest)
	common.ParseHttpBodyParams(request, &requestData)

	// if httpHeader.IsValid() == false {
	if (httpHeader.IsValidTimestamp() == false) ||
		(httpHeader.IsValidTokenhash() == false) ||
		((requestData.Param.Type != LOGIN_PASSWORD) && (requestData.Param.Type != PAYMENT_PASSWORD)) ||
		(len(requestData.Param.Secret) < 1) {
		logger.Info("modify pwd: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("modify pwd: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		logger.Info("modify pwd: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}


	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	// 解码 secret 参数
	// secretString := requestData.Param.Secret
	secret := new(modifySecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	errT := DecryptSecret(requestData.Param.Secret, key, iv, &secret)
	if errT != constants.RC_OK {
		logger.Info("modify pwd: decrypt secret error:", errT)
		response.SetResponseBase(errT)
		return
	}
	if len(secret.NewPwd) < 1 {
		logger.Info("modify pwd: new password is empty, length is:", len(secret.NewPwd))
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	// 检查新旧密码是否重复
	if secret.Pwd == secret.NewPwd {
		logger.Info("modify pwd: orginal password equal to new password in request param")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	account, err := common.GetAccountByUID(uidString)
	if err != nil {
		logger.Info("modify pwd: get account info from DB error:", err)
		response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
		return
	}

	uid := utils.Str2Int64(uidString)

	// 解析出“sha256(密码)”
	// 数据库实际保存的密码格式为“sha256(sha256(密码) + uid)”
	newPwdDb := utils.Sha256(secret.NewPwd + uidString)

	modifyType := requestData.Param.Type
	if modifyType == LOGIN_PASSWORD {
		// 检查密码为空
		if secret.Pwd == "" {
			logger.Info("modify pwd: original pwd is empty")
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		// check old password
		pwdDb := utils.Sha256(secret.Pwd + uidString)
		if account.LoginPassword != pwdDb {
			logger.Info("modify pwd: orginal login password in DB not equal to new password")
			response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
			return
		}
		// save to db
		if err := common.SetLoginPassword(uid, newPwdDb); err != nil {
			logger.Info("modify pwd: save new login pwd to DB error", err)
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		}

	} else if modifyType == PAYMENT_PASSWORD {
		if secret.Pwd == "" {
			// first time of set payment password
			// 检查交易密码是否被设置过
			if account.PaymentPassword != "" {
				logger.Info("modify pwd: you have set payment pwd before")
				response.SetResponseBase(constants.RC_PARAM_ERR)
				return
			}
		} else {
			// check old password
			pwdDb := utils.Sha256(secret.Pwd + uidString) //原密码哈希，首次设置交易密码此处为空
			if account.PaymentPassword != pwdDb {
				logger.Info("modify pwd: orginal payment password in DB not equal to new password")
				response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
				return
			}
		}
		// save to db
		if err := common.SetPaymentPassword(uid, newPwdDb); err != nil {
			logger.Info("modify pwd: save new payment pwd to DB error", err)
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		}
	}
}
