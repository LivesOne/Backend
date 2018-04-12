package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
	"math"
)

type lockRemoveReqData struct {
	Param lockRemoveParam `json:"param"`
}

type lockRemoveParam struct {
	Secret string `json:"secret"`
}

type lockRemoveSecret struct {
	Id  int64  `json:"id"`
	Pwd string `json:"pwd"`
}

type resData struct {
	Cost string
}

// sendVCodeHandler
type lockRemoveHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *lockRemoveHandler) Method() string {
	return http.MethodPost
}

func (handler *lockRemoveHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: 0, // data expire Int 失效时间，单位秒
	}
	defer common.FlushJSONData2Client(response, writer)

	//requestData := lockRemoveRequest{} // request body
	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		logger.Info("asset lockRemove: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("asset lockRemove: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}

	uid := utils.Str2Int64(uidString)

	requestData := new(lockRemoveReqData)

	if !common.ParseHttpBodyParams(request, requestData) {
		response.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	secret := new(lockRemoveSecret)
	err :=	decodeAssetLockSecret(requestData.Param.Secret, key, iv,secret)

	if err != nil {
		logger.Error("decode secret error",err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	pwd := secret.Pwd
	if !common.CheckPaymentPwd(uid, pwd) {
		response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
		return
	}

	al := common.QueryAssetLock(secret.Id)

	if al == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	penaltyMoney := CalculationPenaltyMoney(al)

	txid := common.GenerateTxID()

	if txid == -1 {
		logger.Error("txid is -1  ")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}


	if ok,e := common.RemoveAssetLock(txid,al,penaltyMoney);ok {
		response.Data = resData{
			Cost: utils.LVTintToFloatStr(penaltyMoney),
		}
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


func CalculationPenaltyMoney(al *common.AssetLock)int64{
	//获取当前时间戳
	ts := utils.GetTimestamp13()
	//计算剩余毫秒
	lts := al.End - ts
	//剩余毫秒除以月毫秒数，向上取整 计算出m
	m := math.Ceil(float64(lts)/constants.ASSET_LOCK_MONTH_TIMESTAMP)

	t := float64(al.Month)
	//计算系数
	a := 0.5
	//锁仓数 s 从数据库中取出并转换成正常的lvt数
	s := float64(utils.LVTintToNamorInt(al.ValueInt))
	//L=（m/T）*0.5*S
	//计算后得出的lvt数为float 需要转换成数据库存储的格式
	l := utils.NamorFloatToLVTint(m/t*a*s)
	return l
}
