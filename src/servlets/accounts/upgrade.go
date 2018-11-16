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
	WxCode  string `json:"wx_code"`
	AppType string `json:"app_type"`
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

			if f, e := common.SecondAuthWX(uid, secret.AppType, secret.WxCode); !f {
				response.SetResponseBase(e)
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
