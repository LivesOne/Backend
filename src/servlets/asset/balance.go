package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
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

type balanceDetial struct {
	Currency   string `json:"currency"`
	Balance    string `json:"balance"`
	Locked     string `json:"locked"`
	Income     string `json:"income"`
	Lastmodify int64  `json:"lastmodify"`
	Status     int    `json:"status"`
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
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
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

	currencyList := []string{constants.TRADE_CURRENCY_LVT, constants.TRADE_CURRENCY_LVTC, constants.TRADE_CURRENCY_ETH, constants.TRADE_CURRENCY_EOS, constants.TRADE_CURRENCY_BTC}
	response.Data = buildAllBalanceDetail(currencyList, uid)

}

func buildAllBalanceDetail(currencyList []string, uid int64) []balanceDetial {

	bds := make([]balanceDetial, 0)
	for _, v := range currencyList {
		switch v {
		case constants.TRADE_CURRENCY_LVT:
			balance, locked, income, lastmodify, status, err := common.QueryBalance(uid)
			if err != nil {
				logger.Error("query balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		case constants.TRADE_CURRENCY_LVTC:
			balance, locked, income, lastmodify, status, err := common.QueryBalanceLvtc(uid)
			if err != nil {
				logger.Error("query balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		case constants.TRADE_CURRENCY_ETH:
			balance, locked, income, lastmodify, status, err := common.QueryBalanceEth(uid)
			if err != nil {
				logger.Error("query balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		case constants.TRADE_CURRENCY_EOS:
			balance, locked, income, lastmodify, status, err := common.QueryBalanceEos(uid)
			if err != nil {
				logger.Error("query balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceEOSDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		case constants.TRADE_CURRENCY_BTC:
			balance, locked, income, lastmodify, status, err := common.QueryBalanceBtc(uid)
			if err != nil {
				logger.Error("query balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		}
	}

	return bds
}

func buildSingleBalanceDetail(balance, locked, income, lastmodify int64, status int, currency string) balanceDetial {
	return balanceDetial{
		Currency:   currency,
		Balance:    utils.LVTintToFloatStr(balance),
		Locked:     utils.LVTintToFloatStr(locked),
		Income:     utils.LVTintToFloatStr(income),
		Lastmodify: lastmodify,
		Status:     status,
	}
}

func buildSingleBalanceEOSDetail(balance, locked, income, lastmodify int64, status int, currency string) balanceDetial {
	return balanceDetial{
		Currency:   currency,
		Balance:    utils.EOSintToFloatStr(balance),
		Locked:     utils.EOSintToFloatStr(locked),
		Income:     utils.EOSintToFloatStr(income),
		Lastmodify: lastmodify,
		Status:     status,
	}
}
