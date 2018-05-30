package asset

import (
	"net/http"
	"utils/logger"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"database/sql"
	"utils/vcode"
)

type withdrawRequestParams struct {
	AuthType  int    `json:"auth_type"`
	QuotaType int    `json:"quota_type"`
	VcodeType int    `json:"vcode_type"`
	VcodeId   string `json:"vcode_id"`
	Vcode     string `json:"vcode"`
	Secret    string `json:"secret"`
}

type withdrawRequestSecret struct {
	Address string `json:"address"`
	Value   string `json:"value"`
	Pwd     string `json:"pwd"`
}

func (wqs *withdrawRequestSecret) isValid() bool {
	return len(wqs.Address) > 0 && len(wqs.Value) > 0 && len(wqs.Pwd) > 0
}

type withdrawRequestResponse struct {
	Txid string `json:"txid"`
}

type withdrawRequestHandler struct {
}

func (handler *withdrawRequestHandler) Method() string {
	return http.MethodPost
}

func (handler *withdrawRequestHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}

	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)

	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset lockList: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset lockList: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}
	uid := utils.Str2Int64(uidString)

	requestData := withdrawRequestParams{} // request body

	common.ParseHttpBodyParams(request, &requestData)

	if requestData.QuotaType != 1 && requestData.QuotaType != 2 {
		response.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}

	if requestData.VcodeType > 0 {
		acc, err := common.GetAccountByUID(uidString)
		if err != nil && err != sql.ErrNoRows {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		switch requestData.VcodeType {
		case 1:
			if ok, errCode := vcode.ValidateSmsAndCallVCode(acc.Phone, acc.Country, requestData.Vcode, 3600, vcode.FLAG_DEF); !ok {
				log.Info("validate sms code failed")
				response.SetResponseBase(vcode.ConvSmsErr(errCode))
				return
			}
		case 2:
			if ok, resErr := vcode.ValidateSmsUpVCode(acc.Country, acc.Phone, requestData.Vcode); !ok {
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

	secret := new(withdrawRequestSecret)

	if err := utils.DecodeSecret(requestData.Secret, key, iv, secret); err != nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !secret.isValid() {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	//if !validateValue(secret.Value) {
	//	response.SetResponseBase(constants.RC_PARAM_ERR)
	//	return
	//}
	//TODO 检查权限，userLevel中配置
	withdrawAmount := utils.Str2Int64(secret.Value)
	userWithdrawalQuota := common.GetUserWithdrawalQuotaByUid(uid)
	usedWithdrawalQuotaOfCurMonth := common.QueryWithdrawValueOfCurMonth(uid)
	switch requestData.QuotaType {
	case 1:
		if withdrawAmount > userWithdrawalQuota.Day || withdrawAmount > (userWithdrawalQuota.Month - usedWithdrawalQuotaOfCurMonth){
			response.SetResponseBase(constants.RC_INSUFFICIENT_WITHDRAW_QUOTA)
			return
		}
	case 2:
		if withdrawAmount > userWithdrawalQuota.Casual {
			response.SetResponseBase(constants.RC_INSUFFICIENT_WITHDRAW_QUOTA)
			return
		}
	}
	common.Withdraw(uid, withdrawAmount, secret.Address, requestData.QuotaType)
}
