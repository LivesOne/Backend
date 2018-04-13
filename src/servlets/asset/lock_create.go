package asset

import (
	"encoding/json"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type lockCreateReqData struct {
	Param lockCreateParam `json:"param"`
}

type lockCreateParam struct {
	Secret string `json:"secret"`
}

type lockCreateSecret struct {
	Value string `json:"value"`
	Month int    `json:"month"`
	Pwd   string `json:"pwd"`
}

func (lc *lockCreateSecret)Valid()bool{
	return len(lc.Value)>0&&
			lc.Month >0 &&
			len(lc.Pwd) >0 &&
			utils.Str2Int(lc.Value)>0
}

// sendVCodeHandler
type lockCreateHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *lockCreateHandler) Method() string {
	return http.MethodPost
}

func (handler *lockCreateHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	//requestData := lockCreateRequest{} // request body
	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		logger.Info("asset lockCreate: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("asset lockCreate: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		logger.Info("asset lockCreate: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey,httpHeader.Signature,httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}


	requestData := new(lockCreateReqData)

	if !common.ParseHttpBodyParams(request, requestData) {
		response.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}

	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	secret := new(lockCreateSecret)
	err :=	decodeAssetLockSecret(requestData.Param.Secret, key, iv,secret)
	if err != nil {
		logger.Error("decode secret error",err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !secret.Valid() {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}


	uid := utils.Str2Int64(uidString)

	pwd := secret.Pwd
	if !common.CheckPaymentPwd(uid, pwd) {
		response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
		return
	}



	begin := utils.GetTimestamp13()
	//计算结束时间
	end := begin + (int64(secret.Month) * constants.ASSET_LOCK_MONTH_TIMESTAMP)

	assetLock := &common.AssetLock{
		Uid:      uid,
		Value:    secret.Value,
		ValueInt: utils.FloatStrToLVTint(secret.Value),
		Month:    secret.Month,
		Hashrate: getLockHashrate(secret),
		Begin:    begin,
		End:      end,
	}


	if ok, e := common.CreateAssetLock(assetLock); ok {
		response.Data = assetLock
	} else {
		switch e {
		case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
			response.SetResponseBase(constants.RC_INSUFFICIENT_BALANCE)
		case constants.TRANS_ERR_SYS:
			response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
		case constants.TRANS_ERR_ASSET_LIMITED:
			response.SetResponseBase(constants.RC_ACCOUNT_ACCESS_LIMITED)
		}
	}

}


func getLockHashrate(al *lockCreateSecret)int{
	//锁仓数额	B	[用户自定义填充]，锁仓额为1000LVT的倍数
	b:= utils.Str2Float64(al.Value)
	//锁仓期间	T	用户选择：1个月、3个月、6个月、12个月，24个月
	t := float64(al.Month)

	//算力系数 a=0.2 计算算力为整数，a=0.2 扩大100倍 a := 20
	a := float64(20)
	//锁仓算力	S	S=B/100000*T*a*100%（a=0.2）
	s := b/100000*t*a

	//Mmax=500%，大于500%取500%
	//四舍五入后数值大于500 取500
	if re := utils.Round(s);re <= constants.ASSET_LOCK_MAX_VALUE {
		return re
	}
	return constants.ASSET_LOCK_MAX_VALUE
}


func decodeAssetLockSecret(secret, key, iv string,secretPtr interface{}) error {
	if len(secret) == 0 {
		return nil
	}
	logger.Debug("secret ", secret)
	secJson, err := utils.AesDecrypt(secret, key, iv)
	if err != nil {
		logger.Error("aes decode error ", err.Error())
		return err
	}
	logger.Debug("base64 and aes decode secret ", secJson)
	err = json.Unmarshal([]byte(secJson), secretPtr)
	if err != nil {
		logger.Error("json Unmarshal error ", err.Error())
		return err
	}
	return nil

}
