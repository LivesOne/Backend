package config

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
)

type (
	TransferFeeParam struct {
		Currency string `json:"currency"`
	}

	transferFeeRequest struct {
		Base  *common.BaseInfo  `json:"base"`
		Param *TransferFeeParam `json:"param"`
	}

	transferFeeResData struct {
		SingleAmountMin string                  `json:"single_amount_min"`
		DailyAmountMax  string                  `json:"daily_amount_max"`
		Fee             []*common.DtTransferFee `json:"fee"`
	}
)
type transferFeeHandler struct {
}

func (handler *transferFeeHandler) Method() string {
	return http.MethodPost
}

func (handler *transferFeeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := transferFeeRequest{} // request body

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

	transAmount, err := common.QueryTransAmount(param.Currency)
	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	if transAmount == nil {
		response.SetResponseBase(constants.RC_INVALID_CURRENCY)
		return
	}

	transferFees := common.QueryTransFeesList(param.Currency)
	if len(transferFees) == 0 {
		response.SetResponseBase(constants.RC_INVALID_CURRENCY)
		return
	}
	response.Data = transferFeeResData{
		SingleAmountMin: utils.Float642Str(transAmount.SingleAmountMin),
		DailyAmountMax:  utils.Float642Str(transAmount.DailyAmountMax),
		Fee:             transferFees,
	}
}
