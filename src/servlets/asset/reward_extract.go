package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
	"servlets/accounts"
)

type rewardExtractSecret struct {
	WxCode string `json:"wx_code"`
}

type rewardExtractParam struct {
	Currency string `json:"currency"`
	Secret   string `json:"secret"`
}

type rewardExtractRequest struct {
	Param *rewardExtractParam `json:"param"`
}

type rewardExtractResData struct {
	Currency string `json:"currency"`
	Income   string `json:"income"`
}

// rewardExtractHandler
type rewardExtractHandler struct {
}

func (handler *rewardExtractHandler) Method() string {
	return http.MethodPost
}

func (handler *rewardExtractHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true, "extract income")
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)

	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset reward extract: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)

	log.Info("user login cache token-hash",httpHeader.TokenHash,"uid",uidString,"key",aesKey)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset reward extract: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset reward extract: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	log.Info("uid", uidString)

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		log.Error("asset reward extract: validate sign failed")
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidString)
	if !common.CheckCreditScore(uid, common.DEF_SCORE) {
		log.Info("asset reward extract: permission denied")
		response.SetResponseBase(constants.RC_PERMISSION_DENIED)
		return
	}

	requestData := new(rewardExtractRequest)

	if !common.ParseHttpBodyParams(request, requestData) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if requestData.Param == nil {
		log.Error("asset reward extract: requestData.Param is nil")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	//判断有合法参数才进行微信二次校验
	if len(requestData.Param.Secret) > 0 {
		// 解码 secret 参数
		secretString := requestData.Param.Secret
		secret := new(rewardExtractSecret)
		iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
		if err := accounts.DecryptSecret(secretString, key, iv, secret); err != constants.RC_OK {
			response.SetResponseBase(err)
			return
		}

		if len(secret.WxCode) > 0 {
			// 微信二次验证，未绑定返回验提取失败
			openId, unionId, _, _, _ := common.GetUserExtendByUid(uid)
			if len(openId) == 0 || len(unionId) == 0 {
				log.Error("asset reward extract: user is not bind wx")
				response.SetResponseBase(constants.RC_WX_SEC_AUTH_FAILED)
				return
			}
			//微信认证并比对id
			if ok, res := common.AuthWX(secret.WxCode); ok {
				if res.Unionid != unionId || res.Openid != openId {
					log.Error("asset reward extract: user check sec wx failed")
					log.Error("asset reward extract: db openId,unionId [", openId, unionId, "]")
					log.Error("asset reward extract: wx result openId,unionId [", res.Openid, res.Unionid, "]")

					response.SetResponseBase(constants.RC_WX_SEC_AUTH_FAILED)
					return
				}
			} else {
				log.Error("asset reward extract: wx auth failed")
				response.SetResponseBase(constants.RC_INVALID_WX_CODE)
				return
			}
		}
	}

	currency := common.CURRENCY_LVTC
	if len(requestData.Param.Currency) > 0 {
		if requestData.Param.Currency != common.CURRENCY_LVTC && requestData.Param.Currency != common.CURRENCY_ETH {
			log.Error("asset reward extract: requestData.Param.Currency is", requestData.Param.Currency)
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		currency = requestData.Param.Currency
	}

	var income int64
	var ok  = true
	var err error
	if currency == common.CURRENCY_ETH {
		_, _, income,_,_, err = common.QueryBalanceEth(uid)
		if err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		if income > 0 {
			ok = common.ExtractIncomeEth(uid, income)
		}
	} else {
		_, _, income,_,_, err = common.QueryBalanceLvtc(uid)
		if err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		if income > 0 {
			ok = common.ExtractIncomeLvtc(uid, income)
		}
	}

	if ok {
		response.Data = rewardExtractResData{
			Currency: currency,
			Income:   utils.LVTintToFloatStr(income),
		}
		log.Info("extract success")
	} else {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		log.Info("extract failed")
	}

}
