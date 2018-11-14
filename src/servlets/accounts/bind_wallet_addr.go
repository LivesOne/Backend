package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
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
	uidStr, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidStr)
	walletAddress := strings.ToLower(requestData.Param.Address)

	if validateWalletAddress(walletAddress) {
		if !strings.HasPrefix(walletAddress, "0x") {
			walletAddress = "0x" + walletAddress
		}
		// 查询是否绑定当前地址
		if err := common.InsertUserWalletAddr(uid, walletAddress); err != nil {
			if err == constants.WALLET_DUP_BIND {
				response.SetResponseBase(constants.RC_DUP_WALLET_ADDRESS)
				return
			}
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
	} else {
		response.SetResponseBase(constants.RC_INVALID_WALLET_ADDRESS_FORMAT)
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
