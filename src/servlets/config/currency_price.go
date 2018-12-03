package config

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strings"
)

type currencyPriceParam struct {
	Currency string `json:"currency"`
}

type currencyPriceRequest struct {
	Base  *common.BaseInfo    `json:"base"`
	Param *currencyPriceParam `json:"param"`
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

	currencyPair := strings.ToUpper(param.Currency)

	if f,data := common.GetCurrencyPrice(currencyPair);f {
		response.Data = data
	}else{
		response.SetResponseBase(constants.RC_INVALID_CURRENCY)
	}
}

