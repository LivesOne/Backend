package config

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strings"
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
		Currency []common.CurrencyPricCacheData `json:"currency"`
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
	for _, v := range param.Currency {
		currencyPair := strings.ToUpper(v)
		if f,data := common.GetCurrencyPrice(currencyPair);f {
			batchCurrency.Currency = append(batchCurrency.Currency, *data)
		}
	}
	if len(batchCurrency.Currency) > 0 {
		response.Data = batchCurrency
	} else {
		response.SetResponseBase(constants.RC_INVALID_CURRENCY)
	}


}
