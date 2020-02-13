package accounts

import (
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"golang.org/x/net/context"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"strings"
	"utils"
)

type (
	reChargeAddrParam struct {
		Currency string `json:"currency"`
	}

	reChargeAddrRequest struct {
		// Base  common.BaseInfo `json:"base"`
		Param reChargeAddrParam `json:"param"`
	}

	reChargeAddrRespData struct {
		Currency string `json:"currency,omitempty"`
		Address  string `json:"address,omitempty"`
	}

	hotWalletResult struct {
		Address string `json:"address,omitempty"`
	}
	hotWalletRes struct {
		Code   int             `json:"code"`
		Result hotWalletResult `json:"result,omitempty"`
	}
)

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
	uidStr, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidStr)

	cli := rpc.GetWalletClient()
	if cli == nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	req := &microuser.GetRechargeAddressReq{
		Uid:      uid,
		Currency: currency,
	}

	resp, err := cli.GetRechargeAddress(context.Background(), req)

	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if resp.Result != microuser.ResCode_OK {
		response.SetResponseBase(constants.RC_INVALID_CURRENCY)
		return
	}

	respData := new(reChargeAddrRespData)
	respData.Currency = currency
	respData.Address = resp.WalletAddress
	response.Data = respData
	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
