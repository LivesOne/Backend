package config

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strings"
	"utils"
	"utils/logger"
)

type currencyPriceParam struct {
	Currency string `json:"currency"`
}

type currencyPriceRequest struct {
	Base  *common.BaseInfo    `json:"base"`
	Param *currencyPriceParam `json:"param"`
}

type currencyPriceResData struct {
	Currency string `json:"currency"`
	Current  string `json:"current"`
	Average  string `json:"average"`
}

type currencyPriceHandler struct {
}

func (handler *currencyPriceHandler) Method() string {
	return http.MethodPost
}

func (handler *currencyPriceHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := currencyPriceRequest{} // request body

	parseFlag := common.ParseHttpBodyParams(request, &requestData)
	if !parseFlag {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	base := requestData.Base
	param := requestData.Param
	if base == nil || !base.App.IsValid() || param == nil || len(param.Currency) == 0 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	currencyPair := strings.Split(param.Currency, ",")
	if len(currencyPair) != 2 {
		logger.Warn("currency must be in pair")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	currency := strings.ToUpper(currencyPair[0])
	currency2 := strings.ToUpper(currencyPair[1])

	currentPrice, averagePrice, err := common.QueryCurrencyPrice(currency, currency2)
	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	if currentPrice == "" && averagePrice == "" {
		response.SetResponseBase(constants.RC_INVALID_CURRENCY)
		return
	}
	if strings.Index(currentPrice, ",") >= 0 {
		currentPrice = strings.Replace(currentPrice, ",", "", -1)
	}
	if strings.Index(averagePrice, ",") >= 0 {
		averagePrice = strings.Replace(averagePrice, ",", "", -1)
	}
	response.Data = currencyPriceResData{
		Currency: param.Currency,
		Current:  utils.Scientific2Str(currentPrice),
		Average:  utils.Scientific2Str(averagePrice),
	}
}
