package asset

import (
	"database/sql"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"strings"
	"utils"
	"utils/config"
	"utils/logger"
	"utils/vcode"
)

type transPrepareParam struct {
	TxType    int    `json:"tx_type"`
	AuthType  int    `json:"auth_type"`
	VcodeType int    `json:"vcode_type"`
	VcodeId   string `json:"vcode_id"`
	Vcode     string `json:"vcode"`
	Secret    string `json:"secret"`
}

type transPrepareSecret struct {
	To    string `json:"to"`
	Value string `json:"value"`
	Pwd   string `json:"pwd"`
	BizContent map[string]string `json:"biz_content"`
}

func (tps *transPrepareSecret) isValid() bool {
	return len(tps.To) > 0 && len(tps.Value) > 0 && len(tps.Pwd) > 0
}

type transPrepareRequest struct {
	Base  *common.BaseInfo   `json:"base"`
	Param *transPrepareParam `json:"param"`
}

type transPrepareResData struct {
	Txid string `json:"txid"`
}

// sendVCodeHandler
type transPrepareHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *transPrepareHandler) Method() string {
	return http.MethodPost
}

func (handler *transPrepareHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := transPrepareRequest{} // request body

	common.ParseHttpBodyParams(request, &requestData)

	if requestData.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset trans prepare: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset trans prepare: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset trans prepare: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	// vcodeType 大于0的时候开启短信验证 1下行，2上行
	if requestData.Param.VcodeType > 0 {
		acc, err := common.GetAccountByUID(uidString)
		if err != nil && err != sql.ErrNoRows {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		switch requestData.Param.VcodeType {
		case 1:
			if ok, errCode := vcode.ValidateSmsAndCallVCode(acc.Phone, acc.Country, requestData.Param.Vcode, 3600, vcode.FLAG_DEF); !ok {
				log.Info("validate sms code failed")
				response.SetResponseBase(vcode.ConvSmsErr(errCode))
				return
			}
		case 2:
			if ok, resErr := vcode.ValidateSmsUpVCode(acc.Country, acc.Phone, requestData.Param.Vcode); !ok {
				log.Info("validate up sms code failed")
				response.SetResponseBase(resErr)
				return
			}
		default:
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
	}

	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	secret := new(transPrepareSecret)

	if err := utils.DecodeSecret(requestData.Param.Secret, key, iv, secret); err != nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !secret.isValid() {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !validateValue(secret.Value) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	from := utils.Str2Int64(uidString)
	to := utils.Str2Int64(secret.To)
	txType := requestData.Param.TxType

	//chacke to
	switch txType {
	case constants.TX_TYPE_BUY_COIN_CARD:
		to = config.GetWithdrawalConfig().WithdrawalCardEthAcceptAccount // 手续费收款账号
	default:
		//不能给自己转账，不能转无效用户
		if from == to || !common.ExistsUID(to) {
			response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
			return
		}
	}

	log.Debug(from,to,txType)
	//交易类型 只支持，红包，转账，购买，退款 不支持私募，工资
	switch txType {
	case constants.TX_TYPE_TRANS:

		//目标账号非系统账号才校验额度
		if !config.GetConfig().CautionMoneyIdsExist(to) {

			//在转账的情况下，目标为非系统账号，要校验目标用户是否有收款权限，交易员不受收款权限限制
			transLevelOfTo := common.GetTransLevel(to)
			if transLevelOfTo == 0 && !common.CanBeTo(to) {
				response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
				return
			}

			//金额校验不通过，删除pending
			level := common.GetTransLevel(from)
			if f, e := common.CheckAmount(from, utils.FloatStrToLVTint(secret.Value), level); !f {
				response.SetResponseBase(e)
				return
			}
			//校验用户的交易限制
			if f, e := common.CheckPrepareLimit(from, level); !f {
				response.SetResponseBase(e)
				return
			}
		}
	case constants.TX_TYPE_ACTIVITY_REWARD: //如果是活动领取，需要校验转出者的id
		if utils.Str2Float64(secret.Value) > float64(config.GetConfig().MaxActivityRewardValue) {
			response.SetResponseBase(constants.RC_TRANS_AUTH_FAILED)
			return
		}

		if !common.CheckTansTypeFromUid(from, txType) {
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
	case constants.TX_TYPE_BUY:
		if !common.CheckTansTypeFromUid(to, txType) {
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
		//直接放行
	case constants.TX_TYPE_REFUND:
		if !common.CheckTansTypeFromUid(from, txType) {
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
		//直接放行
	case constants.TX_TYPE_THREAD_IN:
		if !common.CheckTansTypeFromUid(to, txType) {
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
		//直接放行
	case constants.TX_TYPE_THREAD_OUT:
		if !common.CheckTansTypeFromUid(from, txType) {
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
		//直接放行
	case constants.TX_TYPE_BUY_COIN_CARD:
		if len(secret.BizContent["quota"]) == 0 {
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	pwd := secret.Pwd
	switch requestData.Param.AuthType {
	case constants.AUTH_TYPE_LOGIN_PWD:
		if !common.CheckLoginPwd(from, pwd) {
			response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
			return
		}
	case constants.AUTH_TYPE_PAYMENT_PWD:
		if !common.CheckPaymentPwd(from, pwd) {
			response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
			return
		}
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	bizContent :=  utils.ToJSON(secret.BizContent)
	//调用统一提交流程
	if txid, resErr := common.PrepareLVTTrans(from, to, requestData.Param.TxType, secret.Value,bizContent); resErr == constants.RC_OK {
		response.Data = transPrepareResData{
			Txid: txid,
		}
	} else {
		response.SetResponseBase(resErr)
	}

}

func validateValue(value string) bool {
	if utils.Str2Float64(value) > 0 {
		index := strings.Index(value, ".")
		last := value[index+1:]
		if len(last) <= 8 {
			return true
		}
	}
	return false
}
