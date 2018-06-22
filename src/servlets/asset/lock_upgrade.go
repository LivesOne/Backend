package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type lockUpgradeReqData struct {
	Param *lockUpgradeParam `json:"param"`
}

type lockUpgradeParam struct {
	AuthType int    `json:"auth_type"`
	Secret   string `json:"secret"`
}

type lockUpgradeSecret struct {
	Id    string `json:"id"`
	Month int    `json:"month"`
	Pwd   string `json:"pwd"`
}

func (lc *lockUpgradeSecret) Valid() bool {
	return len(lc.Id) > 0 &&
		lc.Month > 0 &&
		len(lc.Pwd) > 0 &&
		utils.Str2Int(lc.Id) > 0
}

// sendVCodeHandler
type lockUpgradeHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *lockUpgradeHandler) Method() string {
	return http.MethodPost
}

func (handler *lockUpgradeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	log := logger.NewLvtLogger(true, "lockUpgradeHandler")
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	//requestData := lockUpgradeRequest{} // request body
	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset lockUpgrade: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset lockUpgrade: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset lockUpgrade: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		log.Info("sign valid failed ")
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	requestData := new(lockUpgradeReqData)

	if !common.ParseHttpBodyParams(request, requestData) {
		log.Error("parse http body params failed")
		response.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}

	reqParam := requestData.Param

	if reqParam == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	secret := new(lockUpgradeSecret)

	if err := utils.DecodeSecret(reqParam.Secret, key, iv, secret); err != nil {
		log.Error("decode secret error", err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !secret.Valid() {
		log.Info("secret valid failed", utils.ToJSON(secret))
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	uid := utils.Str2Int64(uidString)
	//校验用户等级是否可以锁仓
	if !common.CanLockAsset(uid) {
		response.SetResponseBase(constants.RC_USER_LEVEL_LIMIT)
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

	al := common.QueryAssetLock(assetLockId, uid)
	//校验锁仓记录是否可以被提前解锁
	//month > 0 value > 0 end > curr_timestamp
	if al == nil || !al.IsOk() || al.Type == common.ASSET_LOCK_TYPE_DRAW {
		response.SetResponseBase(constants.RC_INVALID_LOCK_ID)
		return
	}

	begin := utils.GetTimestamp13()
	//计算结束时间
	end := begin + (int64(secret.Month) * constants.ASSET_LOCK_MONTH_TIMESTAMP)

	al.Month = secret.Month
	al.Begin = begin
	al.End = end
	al.Type = common.ASSET_LOCK_TYPE_DRAW
	al.Hashrate = getLockHashrate(secret.Month, al.Value)

	if ok, e := common.UpgradeAssetLock(al); ok {
		response.Data = al
	} else {
		switch e {
		case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
			response.SetResponseBase(constants.RC_INVALID_LOCK_VALUE)
		case constants.TRANS_ERR_SYS:
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		case constants.TRANS_ERR_ASSET_LIMITED:
			response.SetResponseBase(constants.RC_ACCOUNT_ACCESS_LIMITED)
		case constants.TRANS_ERR_PARAM:
			response.SetResponseBase(constants.RC_PARAM_ERR)
		}
	}

}
