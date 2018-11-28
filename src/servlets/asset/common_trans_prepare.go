package asset

import (
	"database/sql"
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"servlets/vcode"
	"strings"
	"utils"
	"utils/config"
	"utils/logger"
)

type commonTransPrepareParam struct {
	AuthType  int    `json:"auth_type"`
	VcodeType int    `json:"vcode_type"`
	VcodeId   string `json:"vcode_id"`
	Vcode     string `json:"vcode"`
	Remark    string `json:"remark"`
	Secret    string `json:"secret"`
}

type commonTransPrepareSecret struct {
	To          string `json:"to"`
	Currency    string `json:"currency"`
	Value       string `json:"value"`
	FeeCurrency string `json:"fee_currency"`
	Fee         string `json:"fee"`
	Pwd         string `json:"pwd"`
}

func (tps *commonTransPrepareSecret) isValid() bool {
	if tps.Fee == "" {
		tps.Fee = "0"
	}
	return len(tps.To) > 0 && len(tps.Value) > 0 && len(tps.Currency) > 0 &&
		len(tps.FeeCurrency) > 0 && len(tps.Pwd) > 0
}

type commonTransPrepareRequest struct {
	Base  *common.BaseInfo         `json:"base"`
	Param *commonTransPrepareParam `json:"param"`
}

type commonTransPrepareResData struct {
	Txid     string `json:"txid"`
	Currency string `json:"currency"`
}

// sendVCodeHandler
type commonTransPrepareHandler struct {
}

func (handler *commonTransPrepareHandler) Method() string {
	return http.MethodPost
}

func (handler *commonTransPrepareHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := commonTransPrepareRequest{} // request body

	if ! common.ParseHttpBodyParams(request, &requestData) {
		response.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}

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
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
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
		acc, err := rpc.GetUserInfo(utils.Str2Int64(uidString))
		if err != nil && err != sql.ErrNoRows {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		switch requestData.Param.VcodeType {
		case 1:
			if ok, errCode := vcode.ValidateSmsAndCallVCode(acc.Phone, int(acc.Country), requestData.Param.Vcode, 3600, vcode.FLAG_DEF); !ok {
				log.Info("validate sms code failed")
				response.SetResponseBase(vcode.ConvSmsErr(errCode))
				return
			}
		case 2:
			if ok, resErr := vcode.ValidateSmsUpVCode(int(acc.Country), acc.Phone, requestData.Param.Vcode); !ok {
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

	secret := new(commonTransPrepareSecret)

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
	if from == to || !rpc.UserExists(to) {
		response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
		return
	}

	currency := strings.ToUpper(secret.Currency)
	feeCurrency := strings.ToUpper(secret.FeeCurrency)
	feeTransToAcc := config.GetConfig().TransFeeAccountUid
	//var feeInt int64
	//switch currency {
	//case constants.TRADE_CURRENCY_EOS:
	//	// 校验Eos 日限额及单笔交易额限制
	//	if err := common.VerifyEosTrans(from, secret.Value); err != constants.RC_OK {
	//		response.SetResponseBase(err)
	//		return
	//	}
	//	// 手续费校验
	//	err := common.CheckTransFee(secret.Value, secret.Fee, currency, secret.FeeCurrency)
	//	if err != constants.RC_OK {
	//		response.SetResponseBase(err)
	//		return
	//	}
	//	feeInt = utils.FloatStrToEOSint(secret.Fee)
	//case constants.TRADE_CURRENCY_BTC:
	//	// 校验Btc 日限额及单笔交易额限制
	//	if err := common.VerifyBtcTrans(from, secret.Value); err != constants.RC_OK {
	//		response.SetResponseBase(err)
	//		return
	//	}
	//	// 手续费校验
	//	err := common.CheckTransFee(secret.Value, secret.Fee, currency, secret.FeeCurrency)
	//	if err != constants.RC_OK {
	//		response.SetResponseBase(err)
	//		return
	//	}
	//	feeInt = utils.FloatStrToLVTint(secret.Fee)
	//case constants.TRADE_CURRENCY_ETH:
	//	// 校验ETH 日限额及单笔交易额限制
	//	if err := common.VerifyEthTrans(from, secret.Value); err != constants.RC_OK {
	//		response.SetResponseBase(err)
	//		return
	//	}
	//	// 手续费校验
	//	err := common.CheckTransFee(secret.Value, secret.Fee, currency, secret.FeeCurrency)
	//	if err != constants.RC_OK {
	//		response.SetResponseBase(err)
	//		return
	//	}
	//	feeInt = utils.FloatStrToLVTint(secret.Fee)
	//case constants.TRADE_CURRENCY_LVT:
	//	// 校验LVT 用户每日prepare次数限制及额度限制
	//	if err := common.VerifyLVTTrans(from); err != constants.RC_OK {
	//		response.SetResponseBase(err)
	//		return
	//	}
	//	// lvt 交易员不限制转账额度，不收转账手续费
	//	feeInt = 0
	//case constants.TRADE_CURRENCY_LVTC:
	//	// 校验LVTC 日限额及单笔交易额限制、目标账号收款权限
	//	if err := common.VerifyLVTCTrans(from, secret.Value); err != constants.RC_OK {
	//		response.SetResponseBase(err)
	//		return
	//	}
	//	// 手续费校验
	//	err := common.CheckTransFee(secret.Value, secret.Fee, currency, secret.FeeCurrency)
	//	if err != constants.RC_OK {
	//		response.SetResponseBase(err)
	//		return
	//	}
	//	feeInt = utils.FloatStrToLVTint(secret.Fee)
	//default:
	//	response.SetResponseBase(constants.RC_INVALID_CURRENCY)
	//	return
	//}
	if currency == feeCurrency && from == feeTransToAcc {
		response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
		return
	}

	pwd := secret.Pwd
	switch requestData.Param.AuthType {
	case constants.AUTH_TYPE_LOGIN_PWD:
		if f, _ := rpc.CheckPwd(from, pwd, microuser.PwdCheckType_LOGIN_PWD); !f {
			response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
			return
		}
	case constants.AUTH_TYPE_PAYMENT_PWD:
		if f, _ := rpc.CheckPwd(from, pwd, microuser.PwdCheckType_PAYMENT_PWD); !f {
			response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
			return
		}
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	var txid string
	var resErr constants.Error
	//bizContent := common.TransBizContent{
	//	FeeCurrency: feeCurrency,
	//	Fee:         feeInt,
	//	Remark:      requestData.Param.Remark,
	//}
	////bizContentStr := utils.ToJSON(bizContent)
	//////转账分币种进行
	////switch currency {
	////case constants.TRADE_CURRENCY_EOS:
	////	intValue := utils.FloatStrToEOSint(secret.Value)
	////	txid, _, resErr = common.PrepareTradePending(from, to, intValue, constants.TX_TYPE_TRANS, bizContentStr)
	////case constants.TRADE_CURRENCY_BTC:
	////	intValue := utils.FloatStrToLVTint(secret.Value)
	////	txid, _, resErr = common.PrepareTradePending(from, to, intValue, constants.TX_TYPE_TRANS, bizContentStr)
	////case constants.TRADE_CURRENCY_ETH:
	////	intValue := utils.FloatStrToLVTint(secret.Value)
	////	txid, _, resErr = common.PrepareTradePending(from, to, intValue, constants.TX_TYPE_TRANS, bizContentStr)
	////case constants.TRADE_CURRENCY_LVT:
	////	txid, resErr = common.PrepareLVTTrans(from, to, constants.TX_TYPE_TRANS, secret.Value, bizContentStr, bizContent.Remark)
	////case constants.TRADE_CURRENCY_LVTC:
	////	txid, resErr = common.PrepareLVTCTrans(from, to, constants.TX_TYPE_TRANS, secret.Value, bizContentStr, bizContent.Remark)
	////}
	txid, _, resErr = common.TransferPrepare(from, to, secret.Value, secret.Fee, strings.ToUpper(currency), strings.ToUpper(secret.FeeCurrency), requestData.Param.Remark)
	if resErr == constants.RC_OK {
		response.Data = commonTransPrepareResData{
			Txid:     txid,
			Currency: secret.Currency,
		}
	} else {
		response.SetResponseBase(resErr)
	}
}
