package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type balanceParam struct {
	Uid string `json:"uid"`
}

type balanceRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *balanceParam    `json:"param"`
}

type balanceResData struct {
	Balance     string `json:"balance"`
	Locked      string `json:"locked"`
	LvtcBalance string `json:"lvtc_balance"`
	LvtcLocked  string `json:"lvtc_locked"`
	LvtcIncome  string `json:"lvtc_income"`
	EthBalance  string `json:"eth_balance"`
	EthLocked   string `json:"eth_locked"`
	EthIncome   string `json:"eth_income"`
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
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	//requestData := balanceRequest{} // request body
	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset balance: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset balance: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset balance: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidString)

	balance, locked, err := common.QueryBalance(uid)
	lvtcBalance, lvtcLocked, lvtcIncome, errLvtc := common.QueryBalanceLvtc(uid)
	ethBalance, ethLocked, ethIncome, errEth := common.QueryBalanceEth(uid)
	if err != nil || errLvtc != nil || errEth != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
	} else {
		response.Data = balanceResData{
			Balance:     utils.LVTintToFloatStr(balance),
			Locked:      utils.LVTintToFloatStr(locked),
			LvtcBalance: utils.LVTintToFloatStr(lvtcBalance),
			LvtcLocked:  utils.LVTintToFloatStr(lvtcLocked),
			LvtcIncome:  utils.LVTintToFloatStr(lvtcIncome),
			EthBalance:  utils.LVTintToFloatStr(ethBalance),
			EthLocked:   utils.LVTintToFloatStr(ethLocked),
			EthIncome:   utils.LVTintToFloatStr(ethIncome),
		}
	}

}
