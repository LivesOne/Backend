package config

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strings"
	"utils"
	"utils/logger"
)

type (
	batchCurrencyPriceParam struct {
		Currency []string `json:"currency"`
	}

	batchCurrencyPriceRequest struct {
		Base  *common.BaseInfo         `json:"base"`
		Param *batchCurrencyPriceParam `json:"param"`
	}

	batchCurrencyPriceResData struct {
		Currency []currencyPriceResData `json:"currency"`
	}

	batchCurrencyPriceHandler struct {
	}
)

func (handler *batchCurrencyPriceHandler) Method() string {
	return http.MethodPost
}

func (handler *batchCurrencyPriceHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := batchCurrencyPriceRequest{} // request body

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
	var batchCurrency batchCurrencyPriceResData
	for _, cp := range param.Currency {
		currencyPair := strings.Split(cp, ",")
		if len(currencyPair) != 2 {
			logger.Warn("currency must be in pair")
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
		currency := strings.ToUpper(strings.Trim(currencyPair[0], " "))
		currency2 := strings.ToUpper(strings.Trim(currencyPair[1], " "))
		if len(currency) == 0 || len(currency2) == 0 {
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}

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
		resCurrency := currencyPriceResData{
			Currency: cp,
			Current:  utils.Scientific2Str(currentPrice),
			Average:  utils.Scientific2Str(averagePrice),
		}
		batchCurrency.Currency = append(batchCurrency.Currency, resCurrency)
	}
	response.Data = batchCurrency
}
