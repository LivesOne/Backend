package asset

import (
	"encoding/json"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"strings"
	"utils"
	"utils/config"
	"utils/logger"
)

type transPrepareParam struct {
	TxType   int    `json:"tx_type"`
	AuthType int    `json:"auth_type"`
	Secret   string `json:"secret"`
}

type transPrepareSecret struct {
	To    string `json:"to"`
	Value string `json:"value"`
	Pwd   string `json:"pwd"`
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

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: 0, // data expire Int 失效时间，单位秒
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
		logger.Info("asset trans prepare: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("asset trans prepare: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		logger.Info("asset trans prepare: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	secret := decodeSecret(requestData.Param.Secret, key, iv)

	if secret == nil || !secret.isValid() {
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
	if from == to || !common.ExistsUID(to){
		response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
		return
	}

	txType := requestData.Param.TxType

	//交易类型 只支持，红包，转账，购买，退款 不支持私募，工资
	switch txType {
	case constants.TX_TYPE_TRANS:



		//目标账号非系统账号才校验额度
		if !config.GetConfig().CautionMoneyIdsExist(to) {

			//在转账的情况下，目标为非系统账号，要校验目标用户是否有收款权限
			if !common.CanBeTo(to) {
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
	case constants.TX_TYPE_ACTIVITY_REWARD://如果是活动领取，需要校验转出者的id
		if utils.Str2Float64(secret.Value) > float64(config.GetConfig().MaxActivityRewardValue) {
			response.SetResponseBase(constants.RC_TRANS_AUTH_FAILED)
			return
		}

		if !common.CheckTansTypeFromUid(from, txType) {
			response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
			return
		}
	case constants.TX_TYPE_BUY:
		//直接放行
	case constants.TX_TYPE_REFUND:
		//直接放行
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

	txid := common.GenerateTxID()

	if txid == -1 {
		logger.Error("txid is -1  ")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	txh := common.DTTXHistory{
		Id:     txid,
		Status: constants.TX_STATUS_DEFAULT,
		Type:   requestData.Param.TxType,
		From:   from,
		To:     to,
		Value:  utils.FloatStrToLVTint(secret.Value),
		Ts:     utils.TXIDToTimeStamp13(txid),
		Code:   constants.TX_CODE_SUCC,
	}
	err := common.InsertPending(&txh)
	if err != nil {
		logger.Error("insert mongo db error ", err.Error())
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
	} else {
		response.Data = transPrepareResData{
			Txid: utils.Int642Str(txid),
		}
	}

}

func decodeSecret(secret, key, iv string) *transPrepareSecret {
	if len(secret) == 0 {
		return nil
	}
	logger.Debug("secret ", secret)
	secJson, err := utils.AesDecrypt(secret, key, iv)
	if err != nil {
		logger.Error("aes decode error ", err.Error())
		return nil
	}
	logger.Debug("base64 and aes decode secret ", secJson)
	tps := transPrepareSecret{}
	err = json.Unmarshal([]byte(secJson), &tps)
	if err != nil {
		logger.Error("json Unmarshal error ", err.Error())
		return nil
	}
	return &tps

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
