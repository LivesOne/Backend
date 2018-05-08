package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"servlets/token"
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

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)

	if httpHeader.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}


	requestData := new(upgradeRequest)

	common.ParseHttpBodyParams(request, requestData)
	//判断有合法参数才进行微信二次校验
	if requestData.Param != nil && len(requestData.Param.Secret) >0 {
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
			openId,unionId,_ := common.GetUserExtendByUid(uid)
			if len(openId) == 0 || len(unionId) == 0 {
				response.SetResponseBase(constants.RC_UPGRAD_FAILED)
				return
			}
			//微信认证并比对id
			if ok,res := common.AuthWX(secret.WxCode);ok {
				if res.Unionid != unionId || res.Openid != openId {
					response.SetResponseBase(constants.RC_WX_SEC_AUTH_FAILED)
					return
				}
			} else {
				response.SetResponseBase(constants.RC_INVALID_WX_CODE)
				return
			}

		}

	}

	if ok,level := common.UserUpgrade(uidString);ok {
		response.Data = upgradeResData{
			Level: level,
		}
	} else {
		response.SetResponseBase(constants.RC_UPGRAD_FAILED)
	}





}
