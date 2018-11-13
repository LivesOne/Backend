package config

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
)

type (
	WithdrawalFeeParam struct {
		Currency string `json:"currency"`
	}

	withdrawalFeeRequest struct {
		Base  *common.BaseInfo    `json:"base"`
		Param *WithdrawalFeeParam `json:"param"`
	}

	withdrawalFeeResData struct {
		SingleAmountMin string                    `json:"single_amount_min"`
		DailyAmountMax  string                    `json:"daily_amount_max"`
		Fee             []*common.DtWithdrawalFee `json:"fee"`
	}

	withdrawalFeeHandler struct {
	}
)

func (handler *withdrawalFeeHandler) Method() string {
	return http.MethodPost
}

func (handler *withdrawalFeeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := withdrawalFeeRequest{} // request body
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

	withdrawAmount, err := common.QueryWithdrawalAmount(param.Currency)
	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	if withdrawAmount == nil {
		response.SetResponseBase(constants.RC_INVALID_CURRENCY)
		return
	}

	withdrawalFees := common.QueryWithdrawalFeesList(param.Currency)
	if len(withdrawalFees) == 0 {
		response.SetResponseBase(constants.RC_INVALID_CURRENCY)
		return
	}

	response.Data = withdrawalFeeResData{
		SingleAmountMin: utils.Float642Str(withdrawAmount.SingleAmountMin),
		DailyAmountMax:  utils.Float642Str(withdrawAmount.DailyAmountMax),
		Fee:             withdrawalFees,
	}

}
