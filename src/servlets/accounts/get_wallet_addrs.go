package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
)

type walletAddrParam struct {
	Address string `json:"address"`
	CreateTime int64 `json:"create_time"`
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

	// 查询用户绑定钱包地址
	if addrList, err := common.GetWalletAddrList(uid); err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	} else if addrList != nil {
		var walletList []walletAddrParam
		for _, wallet := range addrList {
			addr := wallet["address"]
			if len(addr) > 0 {
				walletList = append(walletList,
					walletAddrParam{
						Address: wallet["address"],
						CreateTime: utils.Str2Int64(wallet["create_time"]),
					})
			}
		}
		response.Data = walletList
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
