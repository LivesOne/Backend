package asset

import (
	"math"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type lockRemoveReqData struct {
	Param lockRemoveParam `json:"param"`
}

type lockRemoveParam struct {
	AuthType int    `json:"auth_type"`
	Secret   string `json:"secret"`
}

type lockRemoveSecret struct {
	Id  string `json:"id"`
	Pwd string `json:"pwd"`
}

type resData struct {
	Cost string `json:"cost"`
}

func (lc *lockRemoveSecret) Valid() bool {
	return len(lc.Id) > 0 &&
		len(lc.Pwd) > 0
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
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	//requestData := lockRemoveRequest{} // request body
	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset lockRemove: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset lockRemove: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}

	uid := utils.Str2Int64(uidString)

	requestData := new(lockRemoveReqData)

	if !common.ParseHttpBodyParams(request, requestData) {
		response.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}
	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	secret := new(lockRemoveSecret)
	if err := utils.DecodeSecret(requestData.Param.Secret, key, iv, secret); err != nil {
		log.Error("decode secret error", err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !secret.Valid() {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	pwd := secret.Pwd
	switch requestData.Param.AuthType {
	case constants.AUTH_TYPE_LOGIN_PWD:
		if !common.CheckLoginPwd(uid, pwd) {
			response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
			return
		}
	case constants.AUTH_TYPE_PAYMENT_PWD:
		if !common.CheckPaymentPwd(uid, pwd) {
			response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
			return
		}
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	assetLockId := utils.Str2Int64(secret.Id)

	al := common.QueryLvtcAssetLock(assetLockId, uid)
	//校验锁仓记录是否可以被提前解锁
	//month > 0 value > 0 end > curr_timestamp
	if al == nil || !al.IsOk() {
		response.SetResponseBase(constants.RC_INVALID_LOCK_ID)
		return
	}
	// verify is unlock allowed
	if al.AllowUnlock == constants.ASSET_LOCK_UNLOCK_TYPE_ALLOW {
		log.Error("asset unlock not allowed")
		response.SetResponseBase(constants.RC_INVALID_LOCK_ID)
		return
	}
	// verify is currency equal "LVTC"
	if al.Currency != common.CURRENCY_LVTC {
		log.Error("asset unlock currency: ", al.Currency)
		response.SetResponseBase(constants.RC_INVALID_LOCK_ID)
		return
	}

	penaltyMoney := CalculationPenaltyMoney(al)

	log.Info("query asset lock ", utils.ToJSON(al))

	txid := common.GenerateTxID()

	if txid == -1 {
		log.Error("txid is -1  ")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if ok, e := common.RemoveAssetLock(txid, al, penaltyMoney); ok {
		response.Data = resData{
			Cost: utils.LVTintToFloatStr(penaltyMoney),
		}
	} else {
		switch e {
		case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
			response.SetResponseBase(constants.RC_INSUFFICIENT_BALANCE)
		case constants.TRANS_ERR_SYS:
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		case constants.TRANS_ERR_ASSET_LIMITED:
			response.SetResponseBase(constants.RC_ACCOUNT_ACCESS_LIMITED)
		}
	}

}

func CalculationPenaltyMoney(al *common.AssetLockLvtc) int64 {
	//获取当前时间戳
	ts := utils.GetTimestamp13()
	//计算剩余毫秒
	lts := al.End - ts
	//剩余毫秒除以月毫秒数，向上取整 计算出m
	m := math.Ceil(float64(lts) / constants.ASSET_LOCK_MONTH_TIMESTAMP)

	t := float64(al.Month)
	//计算系数
	a := 0.5
	//锁仓数 s 从数据库中取出并转换成正常的lvt数
	s := float64(utils.LVTintToNamorInt(al.ValueInt))
	//L=（m/T）*0.5*S
	//计算后得出的lvt数为float 需要转换成数据库存储的格式
	l := utils.NamorFloatToLVTint(m / t * a * s)
	logger.Info("m", m, "t", t, "a", a, "s", s, "m/t*a*s", l)
	return l
}
