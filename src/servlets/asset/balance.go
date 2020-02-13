package asset

import (
	"fmt"
	"math"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/config"
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
	Currency    string `json:"currency"`
	Balance     string `json:"balance"`
	BalanceLite string `json:"balance_lite"`
	Locked      string `json:"locked"`
	Income      string `json:"income"`
	Lastmodify  int64  `json:"lastmodify"`
	Status      int    `json:"status"`
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

	currencyList := []string{constants.TRADE_CURRENCY_LVT, constants.TRADE_CURRENCY_LVTC, constants.TRADE_CURRENCY_ETH,
		constants.TRADE_CURRENCY_EOS, constants.TRADE_CURRENCY_BTC, constants.TRADE_CURRENCY_BSV}
	response.Data = buildAllBalanceDetail(currencyList, uid)

}

func buildAllBalanceDetail(currencyList []string, uid int64) []balanceDetial {
	fmt.Println(currencyList)
	bds := make([]balanceDetial, 0)
	for _, v := range currencyList {
		switch v {
		case constants.TRADE_CURRENCY_LVT:
			balance, locked, income, lastmodify, status, err := common.QueryBalance(uid)
			if err != nil {
				logger.Error("query lvt balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		case constants.TRADE_CURRENCY_LVTC:
			balance, locked, income, lastmodify, status, err := common.QueryBalanceLvtc(uid)
			if err != nil {
				logger.Error("query lvtc balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		case constants.TRADE_CURRENCY_ETH:
			balance, locked, income, lastmodify, status, err := common.QueryBalanceEth(uid)
			if err != nil {
				logger.Error("query eth balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		case constants.TRADE_CURRENCY_EOS:
			balance, locked, income, lastmodify, status, err := common.QueryBalanceEos(uid)
			if err != nil {
				logger.Error("query eos balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		case constants.TRADE_CURRENCY_BTC:
			balance, locked, income, lastmodify, status, err := common.QueryBalanceBtc(uid)
			if err != nil {
				logger.Error("query btc balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)
		case constants.TRADE_CURRENCY_BSV:
			balance, locked, income, lastmodify, status, err := common.QueryBalanceBsv(uid)
			if err != nil {
				logger.Error("query bsv balance error", err.Error())
				return nil
			}
			bd := buildSingleBalanceDetail(balance, locked, income, lastmodify, status, v)
			bds = append(bds, bd)}
	}
	fmt.Println(bds)
	return bds
}

func buildSingleBalanceDetail(balance, locked, income, lastmodify int64, status int, currency string) balanceDetial {
	b, bl,l, i := getFormatBalanceInfo(currency, balance, locked, income)
	return balanceDetial{
		Currency:    currency,
		Balance:     b,
		BalanceLite: bl,
		Locked:      l,
		Income:      i,
		Lastmodify:  lastmodify,
		Status:      status,
	}
}

func getFormatBalanceInfo(currency string, value,locked,income int64) (string, string,string, string) {
	de := config.GetConfig().GetDecimalsByCurrency(currency)
	zeroRes := utils.IntToFloatStrByDecimal(0, 8, 8)
	if de != nil {
		dbdec := int32(de.DBDecimal)
		l := utils.IntToFloatStrByDecimal(locked, dbdec, dbdec)
		i := utils.IntToFloatStrByDecimal(income, dbdec, dbdec)
		balance := utils.IntToFloatStrByDecimal(value, dbdec, dbdec)
		if de.DBDecimal == de.ShowDecimal {
			return balance, balance,l,i
		} else {
			showDec := getShowDecimal(de.DBDecimal, de.ShowDecimal, value)
			logger.Info("format balance currency",currency,"db decimal",dbdec,"show decimal",showDec)
			return balance, utils.IntToFloatStrByDecimal(value, dbdec, showDec),l,i
		}
	}
	return zeroRes,zeroRes,zeroRes,zeroRes
}



func getShowDecimal(dbDec, showDec int, value int64) int32 {
	minValue := int64(math.Pow10(dbDec-showDec))
	logger.Info("getShowDecimal pow10(",dbDec-showDec,") min value",minValue,"value",value)
	if dbDec > showDec && minValue > value {
		return getShowDecimal(dbDec, showDec+1, value)
	}
	return int32(showDec)
}
