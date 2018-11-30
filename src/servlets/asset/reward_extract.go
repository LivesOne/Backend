package asset

import (
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"net/http"
	"servlets/accounts"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)

type rewardExtractSecret struct {
	WxCode  string `json:"wx_code"`
	AppType string `json:"app_type"`
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
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)

	log.Info("user login cache token-hash", httpHeader.TokenHash, "uid", uidString, "key", aesKey)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
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

	score, _ := rpc.GetUserField(uid, microuser.UserField_CREDIT_SCORE)
	if utils.Str2Int(score) < common.DEF_SCORE {
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
		if f, e := common.SecondAuthWX(uid, secret.AppType, secret.WxCode); !f {
			response.SetResponseBase(e)
			return
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
	var ok = true
	var err error
	if currency == common.CURRENCY_ETH {
		_, _, income, _, _, err = common.QueryBalanceEth(uid)
		if err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		if income > 0 {
			ok = common.ExtractIncomeEth(uid, income)
		}
	} else {
		_, _, income, _, _, err = common.QueryBalanceLvtc(uid)
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
