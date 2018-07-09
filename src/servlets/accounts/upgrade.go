package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type upgradeSecret struct {
	WxCode string `json:"wx_code"`
}

type upgradeParam struct {
	Secret string `json:"secret"`
}

type upgradeRequest struct {
	Param *upgradeParam `json:"param"`
}

type upgradeResData struct {
	Level int `json:"level"`
}

// checkVCodeHandler
type upgradeHandler struct {
}

func (handler *upgradeHandler) Method() string {
	return http.MethodPost
}

func (handler *upgradeHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true, "user upgrade")
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)

	if httpHeader.Timestamp < 1 {
		log.Error("timestamp check failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Error("get cache failed")
		response.SetResponseBase(err)
		return
	}

	log.Info("uid", uidString)

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		log.Error("validate sign failed")
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	requestData := new(upgradeRequest)

	common.ParseHttpBodyParams(request, requestData)
	//判断有合法参数才进行微信二次校验
	if requestData.Param != nil && len(requestData.Param.Secret) > 0 {
		// 解码 secret 参数
		secretString := requestData.Param.Secret
		secret := new(upgradeSecret)
		iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
		if err := DecryptSecret(secretString, key, iv, secret); err != constants.RC_OK {
			response.SetResponseBase(err)
			return
		}

		if len(secret.WxCode) > 0 {
			// 微信二次验证
			uid := utils.Str2Int64(uidString)
			//未绑定返回验升级失败
			openId, unionId, _ := common.GetUserExtendByUid(uid)
			if len(openId) == 0 || len(unionId) == 0 {
				log.Error("user is not bind wx")
				response.SetResponseBase(constants.RC_UPGRAD_FAILED)
				return
			}
			//微信认证并比对id
			if ok, res := common.AuthWX(secret.WxCode); ok {
				if res.Unionid != unionId || res.Openid != openId {
					log.Error("user check sec wx failed")
					log.Error("db openId,unionId [", openId, unionId, "]")
					log.Error("wx result openId,unionId [", res.Openid, res.Unionid, "]")
					//二次验证不通过扣10分
					//deductionCreditScore := 10
					//log.Error("deduction credit score :",deductionCreditScore)
					//common.DeductionCreditScore(uid,deductionCreditScore)

					response.SetResponseBase(constants.RC_WX_SEC_AUTH_FAILED)
					return
				}
			} else {
				log.Error("wx auth failed")
				response.SetResponseBase(constants.RC_INVALID_WX_CODE)
				return
			}

		}

	}

	if ok, level := common.UserUpgrade(uidString); ok {
		response.Data = upgradeResData{
			Level: level,
		}
		log.Info("upgrade success")
	} else {
		response.SetResponseBase(constants.RC_UPGRAD_FAILED)
		log.Info("upgrade failed")
	}

}
