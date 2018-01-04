package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils/logger"
	"servlets/token"
	"utils"
)

type balanceParam struct {
	Uid string `json:"uid"`
}

type balanceRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *balanceParam     `json:"param"`
}

type balanceResData struct {
	Balance string `json:"balance"`
}

// sendVCodeHandler
type balanceHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *balanceHandler) Method() string {
	return http.MethodPost
}

func (handler *balanceHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: 0, // data expire Int 失效时间，单位秒
	}
	defer common.FlushJSONData2Client(response, writer)

	//requestData := balanceRequest{} // request body
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

	uid := utils.Str2Int64(uidString)

	balance := common.QueryBalance(uid)

	response.Data = balanceResData{
		Balance:utils.LVTintToFloatStr(balance),
	}


}
func TokenErr2RcErr(tokenErr int) constants.Error {
	switch tokenErr {
	case constants.ERR_INT_OK:
		return constants.RC_OK
	case constants.ERR_INT_TK_DB:
		return constants.RC_PARAM_ERR
	case constants.ERR_INT_TK_DUPLICATE:
		return constants.RC_PARAM_ERR
	case constants.ERR_INT_TK_NOTEXISTS:
		return constants.RC_PARAM_ERR
	default:
		return constants.RC_SYSTEM_ERR
	}
}