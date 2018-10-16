package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/config"
	"utils/logger"
)


// sendVCodeHandler
type lvt2lvtcCountHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *lvt2lvtcCountHandler) Method() string {
	return http.MethodPost
}

func (handler *lvt2lvtcCountHandler) Handle(request *http.Request, writer http.ResponseWriter) {
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

	uid := utils.Str2Int64(uidString)


	resData := &lvt2lvtcResData{
		Lvt:  "0",
		Lvtc: "0",
	}


	balance,_,_,_,_,err := common.QueryBalance(uid)
	if err != nil {
		log.Error("query mysql error",err.Error())
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	if balance > 0 {
		resData.Lvt = utils.LVTintToFloatStr(balance)
		lvtcBalance := balance/int64(config.GetConfig().LvtcHashrateScale)
		resData.Lvtc = utils.LVTintToFloatStr(lvtcBalance)
	}

	response.Data = resData


}
