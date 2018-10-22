package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"strings"
	"utils"
	"utils/config"
)

type reChargeAddrParam struct {
	Currency string `json:"currency"`
}

type reChargeAddrRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param reChargeAddrParam `json:"param"`
}

type reChargeAddrRespData struct {
	Currency string `json:"currency,omitempty"`
	Address  string `json:"address,omitempty"`
}

// bindEMailHandler
type reChargeAddrHandler struct {
}

func (handler *reChargeAddrHandler) Method() string {
	return http.MethodPost
}

func (handler *reChargeAddrHandler) Handle(
	request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)
	requestData := new(reChargeAddrRequest)
	common.ParseHttpBodyParams(request, requestData)

	if httpHeader.Timestamp < 1 || requestData.Param.Currency == "" {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	currency := strings.ToUpper(requestData.Param.Currency)

	// 判断用户身份
	_, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	// 返回充值地址
	var addr string
	for _, rechargeAddr := range config.GetConfig().ReChargeAddress {
		if rechargeAddr.Currency == currency {
			addr = rechargeAddr.Address
			addr = strings.Trim(addr, " ")
			break
		}
	}
	if addr == "" {
		response.SetResponseBase(constants.RC_INVALID_CURRENCY)
		return
	}
	respData := new(reChargeAddrRespData)
	respData.Currency = currency
	respData.Address = addr
	response.Data = respData
	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
