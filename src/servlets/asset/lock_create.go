package asset

import (
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/config"
	"utils/logger"
)

type lockCreateReqData struct {
	Param *lockCreateParam `json:"param"`
}

type lockCreateParam struct {
	AuthType int    `json:"auth_type"`
	Secret   string `json:"secret"`
}

type lockCreateSecret struct {
	Value string `json:"value"`
	Month int    `json:"month"`
	Pwd   string `json:"pwd"`
}

func (lc *lockCreateSecret) Valid() bool {
	return len(lc.Value) > 0 &&
		lc.Month > 0 &&
		len(lc.Pwd) > 0 &&
		utils.Str2Int(lc.Value) > 0
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

	log := logger.NewLvtLogger(true, "lockCreateHandler")
	defer log.InfoAll()
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
		log.Info("asset lockCreate: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset lockCreate: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset lockCreate: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		log.Info("sign valid failed ")
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	requestData := new(lockCreateReqData)

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

	secret := new(lockCreateSecret)

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
	//if !common.CanLockAsset(uid) {
	//	response.SetResponseBase(constants.RC_USER_LEVEL_LIMIT)
	//	return
	//}

	pwd := secret.Pwd
	switch requestData.Param.AuthType {
	case constants.AUTH_TYPE_LOGIN_PWD:
		if f, _ := rpc.CheckPwd(uid, pwd, microuser.PwdCheckType_LOGIN_PWD); !f {
			response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
			return
		}
	case constants.AUTH_TYPE_PAYMENT_PWD:
		if f, _ := rpc.CheckPwd(uid, pwd, microuser.PwdCheckType_PAYMENT_PWD); !f {
			response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
			return
		}
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	begin := utils.GetTimestamp13()
	//计算结束时间
	end := begin + (int64(secret.Month) * constants.ASSET_LOCK_MONTH_TIMESTAMP)
	lvtScale := config.GetConfig().LvtcHashrateScale
	assetLock := &common.AssetLockLvtc{
		Uid:         uid,
		Value:       secret.Value,
		ValueInt:    utils.FloatStrToLVTint(secret.Value),
		Month:       secret.Month,
		Hashrate:    utils.GetLockHashrate(lvtScale, secret.Month, secret.Value),
		Begin:       begin,
		End:         end,
		Currency:    common.CURRENCY_LVTC,
		AllowUnlock: constants.ASSET_LOCK_UNLOCK_TYPE_DEF,
	}
	if assetLock.ValueInt < 100 {
		log.Info("asset create lock less than 100")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if ok, e := common.CreateAssetLock(assetLock); ok {
		response.Data = assetLock
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
