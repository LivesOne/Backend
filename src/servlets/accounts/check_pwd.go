package accounts

import (
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)

const (
	CHECK_TYPE_LOGIN_PWD = 1
	CHECK_TYPE_PAYMENT_PWD = 2

)

type (
	checkPWDRequest struct {
		Param *struct {
			Type   int    `json:"type"`
			Secret string `json:"secret"`
		} `json:"param"`
	}
	checkSecret struct {
		Pwd string `json:"pwd"`
	}
	checkPwdHandler struct {
	}
)

func (handler *checkPwdHandler) Method() string {
	return http.MethodPost
}

func (handler *checkPwdHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	log := logger.NewLvtLogger(true,"checkPwdHandler")
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	if header.Timestamp < 1 {
		log.Error("timestamp check failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(header.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Error("get cache failed")
		response.SetResponseBase(err)
		return
	}

	log.Info("uid", uidString)

	if !utils.SignValid(aesKey, header.Signature, header.Timestamp) {
		log.Error("validate sign failed")
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	requestData := new(checkPWDRequest)
	common.ParseHttpBodyParams(request, requestData)

	if requestData.Param == nil || len(requestData.Param.Secret) == 0 {
		log.Error("wrong check type")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 解码 secret 参数
	secretString := requestData.Param.Secret
	secret := new(checkSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	if err := DecryptSecret(secretString, key, iv, secret); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}
	uid := utils.Str2Int64(uidString)
	var checkType microuser.PwdCheckType
	switch requestData.Param.Type {
	case CHECK_TYPE_LOGIN_PWD:
		checkType = microuser.PwdCheckType_LOGIN_PWD
	case CHECK_TYPE_PAYMENT_PWD:
		checkType = microuser.PwdCheckType_PAYMENT_PWD
	default:
		log.Error("wrong check type")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	flag,err := rpc.CheckPwd(uid,secret.Pwd,checkType)
	if err != nil {
		log.Error("rpc error",err.Error())
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !flag {
		switch requestData.Param.Type {
		case CHECK_TYPE_LOGIN_PWD:
			response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		case CHECK_TYPE_PAYMENT_PWD:
			response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
		}
	}

}


