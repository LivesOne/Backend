package config

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strings"
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

	if data,err := QueryCurrencyPrice(param.Currency);err == constants.RC_OK {
		response.Data = data
	}else{
		response.SetResponseBase(err)
	}
}


func QueryCurrencyPrice(currencyPiar string)(*currencyPriceResData,constants.Error){
	currencyPair := strings.Split(currencyPiar, ",")
	if len(currencyPair) != 2 {
		logger.Warn("currency must be in pair")
		return nil,constants.RC_PARAM_ERR
	}
	currency := strings.ToUpper(strings.TrimSpace(currencyPair[0]))
	currency2 := strings.ToUpper(strings.TrimSpace(currencyPair[1]))

	currentPrice, averagePrice, err := common.QueryCurrencyPrice(currency, currency2)
	if err != nil {
		return nil,constants.RC_SYSTEM_ERR

	}
	if currentPrice == "" && averagePrice == "" {
		logger.Info("currency piar ",currentPrice,averagePrice,"")
		return nil,constants.RC_INVALID_CURRENCY
	}
	//if strings.Index(currentPrice, ",") >= 0 {
	//	currentPrice = strings.Replace(currentPrice, ",", "", -1)
	//}
	//if strings.Index(averagePrice, ",") >= 0 {
	//	averagePrice = strings.Replace(averagePrice, ",", "", -1)
	//}
	return &currencyPriceResData{
		Currency: currencyPiar,
		Current:  currentPrice,
		Average:  averagePrice,
	},constants.RC_OK
}