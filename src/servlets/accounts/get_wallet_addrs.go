package accounts

import (
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"golang.org/x/net/context"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)

type walletAddrParam struct {
	Address    string `json:"address"`
	CreateTime int64  `json:"create_time"`
}

type walletAddrHandler struct {
}

func (handler *walletAddrHandler) Method() string {
	return http.MethodPost
}

func (handler *walletAddrHandler) Handle(
	request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	//response.Data = new(map[string]string)
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)

	if httpHeader.Timestamp < 1 {
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

	cli := rpc.GetWalletClient()
	if cli == nil {
		logger.Error("can not get Wallet rpc client ")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	req := &microuser.UserIdReq{
		Uid: uid,
	}
	resp, err := cli.QueryWallet(context.Background(), req)
	if err != nil {
		logger.Error("rpc query wallet error", err.Error())
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	if resp.Result == microuser.ResCode_OK {
		if len(resp.Wallets) > 0 {
			var walletList []walletAddrParam
			for _, wallet := range resp.Wallets {
				if len(wallet.Address) > 0 {
					walletList = append(walletList,
						walletAddrParam{
							Address:    wallet.Address,
							CreateTime: wallet.CreateTime,
						})
				}
			}
			response.Data = walletList
		}
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
