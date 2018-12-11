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

type bindWalletAddrParam struct {
	Address string `json:"address"`
}

type bindWalletAddrRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param bindWalletAddrParam `json:"param"`
}

type bindWalletAddrHandler struct {
}

func (handler *bindWalletAddrHandler) Method() string {
	return http.MethodPost
}

func (handler *bindWalletAddrHandler) Handle(
	request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)
	requestData := new(bindWalletAddrRequest)
	common.ParseHttpBodyParams(request, requestData)

	if httpHeader.Timestamp < 1 || requestData.Param.Address == "" {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidStr, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if tokenErr != microuser.ResCode_OK {
		response.SetResponseBase(rpc.TokenErr2RcErr(tokenErr))
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidStr)
	addr := strings.ToLower(requestData.Param.Address)

	cli := rpc.GetWalletClient()
	if cli == nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	req := &microuser.WalletBindReq{
		Uid:     uid,
		Address: addr,
	}
	resp, err := cli.BindWallet(context.Background(), req)
	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if resp.Result != microuser.ResCode_OK {
		switch resp.Result {
		case microuser.ResCode_ERR_DUP_DATA:
			response.SetResponseBase(constants.RC_DUP_WALLET_ADDRESS)
			return
		default:
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
	}
	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
