package asset

import (
	"database/sql"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
	"servlets/vcode"
)

type ethTransPrepareParam struct {
	TxType    int    `json:"tx_type"`
	AuthType  int    `json:"auth_type"`
	VcodeType int    `json:"vcode_type"`
	VcodeId   string `json:"vcode_id"`
	Vcode     string `json:"vcode"`
	Secret    string `json:"secret"`
}

type ethTransPrepareSecret struct {
	To         string            `json:"to"`
	Value      string            `json:"value"`
	Pwd        string            `json:"pwd"`
	BizContent map[string]string `json:"biz_content"`
}

func (tps *ethTransPrepareSecret) isValid() bool {
	return len(tps.To) > 0 && len(tps.Value) > 0 && len(tps.Pwd) > 0
}

type ethTransPrepareRequest struct {
	Base  *common.BaseInfo      `json:"base"`
	Param *ethTransPrepareParam `json:"param"`
}

type ethTransPrepareResData struct {
	TradeNo string `json:"trade_no"`
}

// sendVCodeHandler
type ethTransPrepareHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *ethTransPrepareHandler) Method() string {
	return http.MethodPost
}

func (handler *ethTransPrepareHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := ethTransPrepareRequest{} // request body

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

	secret := new(ethTransPrepareSecret)

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

	//不能给自己转账，不能转无效用户
	if from == to || !common.ExistsUID(to) {
		response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
		return
	}

	txType := requestData.Param.TxType

	switch txType {
	//case constants.TX_TYPE_BUY_COIN_CARD:
	//交易类型 23 购买提币卡
	//	if len(secret.BizContent["quota"]) == 0 {
	//		response.SetResponseBase(constants.RC_PARAM_ERR)
	//		return
	//	}
	case constants.TX_TYPE_TRANS:
		//交易类型 4 转账
		secret.BizContent = nil
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

	//调用统一提交流程
	bizContent := utils.ToJSON(secret.BizContent)
	valueInt := utils.FloatStrToLVTint(secret.Value)
	if _, tradeNo, resErr := common.PrepareTradePending(from, to, valueInt, requestData.Param.TxType, bizContent); resErr == constants.RC_OK {
		response.Data = ethTransPrepareResData{
			TradeNo: tradeNo,
		}
	} else {
		response.SetResponseBase(resErr)
	}

}
