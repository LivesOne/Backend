package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"servlets/token"
	"utils/logger"
	"encoding/json"
	"strings"
)



type transPrepareParam struct {
	TxType   int    `json:"tx_type"`
	AuthType int    `json:"auth_type"`
	Secret string `json:"secret"`
}

type transPrepareSecret struct {
	To       string `json:"to"`
	Value    string `json:"value"`
	Pwd      string `json:"pwd"`
}

func (tps *transPrepareSecret)isValid()bool{
	return len(tps.To)>0 && len(tps.Value) > 0 && len(tps.Pwd) >0
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


	if requestData.Base == nil || requestData.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}



	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if  !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash()  {
		logger.Info("modify pwd: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("asset balance: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		logger.Info("asset balance: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]



	secret := decodeSecret(requestData.Param.Secret,key,iv)

	if secret == nil || !secret.isValid() {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !validateValue(secret.Value) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	pwd := secret.Pwd

	from := utils.Str2Int64(uidString)
	switch requestData.Param.AuthType {
	case constants.AUTH_TYPE_LOGIN_PWD:
		if !common.CheckLoginPwd(from,pwd) {
			response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
			return
		}
	case constants.AUTH_TYPE_PAYMENT_PWD:
		if !common.CheckPaymentPwd(from,pwd) {
			response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
			return
		}
	}





	txid := common.GenerateTxID()

	if txid == -1 {
		logger.Error("txid is -1  ")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	to := utils.Str2Int64(secret.To)

	switch requestData.Param.TxType {
		case constants.TX_TYPE_TRANS:
			ts := utils.GetTimestamp13()
			txh := common.DTTXHistory{
				Id:     txid,
				Status: constants.TX_STATUS_DEFAULT,
				Type:   constants.TX_TYPE_TRANS,
				From:   from,
				To:     to,
				Value:  utils.FloatStrToLVTint(secret.Value),
				Ts:     ts,
				Code:   constants.TX_CODE_SUCC,
			}
			err := common.InsertPending(&txh)
			if err != nil {
				logger.Error("insert mongo db error ",err.Error())
				response.SetResponseBase(constants.RC_SYSTEM_ERR)
			} else {
				response.Data = transPrepareResData{
					Txid: utils.Int642Str(txid),
				}
			}
		default:
			response.SetResponseBase(constants.RC_PARAM_ERR)
	}

}


func decodeSecret(secret,key ,iv string)*transPrepareSecret{
	if len(secret) == 0 {
		return nil
	}
	logger.Debug("secret ",secret)
	secJson,err := utils.AesDecrypt(secret,key,iv)
	if err !=nil{
		logger.Error("aes decode error ",err.Error())
		return nil
	}
	logger.Debug("base64 and aes decode secret ",secJson)
	tps := transPrepareSecret{}
	err = json.Unmarshal([]byte(secJson),&tps)
	if err != nil {
		logger.Error("json Unmarshal error ",err.Error() )
		return nil
	}
	return &tps

}


func validateValue(value string)bool {
	if utils.Str2Float64(value) > 0 {
		index := strings.Index(value,".")
		last := value[index+1:]
		if len(last) <= 8 {
			return true
		}
	}
	return false
}