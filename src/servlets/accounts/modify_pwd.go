package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
)

const (
	LOGIN_PASSWORD = 1
	PAYMENT_PASSWORD = 2
)

type modifyPwdParam struct {
	Type int `json:"type"`
	Secret string `json:"secret"`
}

type modifySecret struct {
	Pwd string `json:"pwd"`
	NewPwd string `json:"new_pwd"`
}

type modifyPwdRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param modifyPwdParam `json:"param"`
}

// modifyPwdHandler
type modifyPwdHandler struct {
	header      *common.HeaderParams // request header param
	requestData *modifyPwdRequest    // request body
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

	if httpHeader.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
	}
	uid := utils.Str2Int64(uidString)

	// 解码 secret 参数
	secretString := requestData.Param.Secret
	secret := new(modifySecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	if err := DecryptSecret(secretString, key, iv, &secret); err != constants.RC_OK {
		response.SetResponseBase(err)
	}

	// 解析出“sha256(密码)”
	// 数据库实际保存的密码格式为“sha256(sha256(密码) + uid)”
	pwdDb := utils.Sha256(secret.Pwd + uidString)
	newPwdDb := utils.Sha256(secret.NewPwd + uidString)

	// 判断各种参数不合法的情况
	if secret.NewPwd == "" {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	modifyType := requestData.Param.Type
	if modifyType == LOGIN_PASSWORD {
		// 检查密码为空
		if secret.Pwd == "" {
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		// 检查新旧密码是否重复
		if secret.Pwd == secret.NewPwd {
			response.SetResponseBase(constants.RC_DUP_LOGIN_PWD)
			return
		}
		// check old password
		account, err := common.GetAccountByUID(uidString)
		if err != nil {
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
		if account.LoginPassword != pwdDb {
			response.SetResponseBase(constants.RC_DUP_LOGIN_PWD)
			return
		}
		// save to db
		if err := common.SetLoginPassword(uid, newPwdDb); err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		}
		// send response
		response.SetResponseBase(constants.RC_OK)
		return

	} else if modifyType == PAYMENT_PASSWORD {
		if secret.Pwd == "" {
			// 检查交易密码是否被设置过
			account, err := common.GetAccountByUID(uidString)
			if err != nil {
				response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
				return
			}
			if account.PaymentPassword != "" {
				response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
				return
			}
		} else {
			// check old password
			if secret.NewPwd != secret.Pwd {
				response.SetResponseBase(constants.RC_DUP_PAYMENT_PWD)
				return
			}
			account, err := common.GetAccountByUID(uidString)
			if err != nil {
				response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
				return
			}
			if account.PaymentPassword != pwdDb {
				response.SetResponseBase(constants.RC_DUP_PAYMENT_PWD)
				return
			}
		}
		// save to db
		if err := common.SetPaymentPassword(uid, newPwdDb); err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		}
		// send response
		response.SetResponseBase(constants.RC_OK)
		return

	} else {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
}
