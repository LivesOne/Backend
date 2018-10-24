package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strings"
	"utils"
)

// sendVCodeHandler
type commonTransResultHandler struct {
}

func (handler *commonTransResultHandler) Method() string {
	return http.MethodPost
}

func (handler *commonTransResultHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := commonTransCommitRequest{} // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &requestData)

	if requestData.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if len(requestData.Param.Txid) > 0 && len(requestData.Param.Currency) > 0 {
		txid := utils.Str2Int64(requestData.Param.Txid)
		currency := strings.ToUpper(requestData.Param.Currency)
		switch currency {
		case constants.TRADE_CURRENCY_EOS:
			//数据库存在 返回成功
			if common.CheckEosHistoryByTxid(txid) {
				return
			}
			//pending存在 返回处理中
			if common.CheckTradePendingByTxid(txid) {
				response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
				return
			}
			//都未查到，返回无效的txid
			response.SetResponseBase(constants.RC_INVALID_TXID)
			return
		case constants.TRADE_CURRENCY_BTC:
			//数据库存在 返回成功
			if common.CheckBtcHistoryByTxid(txid) {
				return
			}
			//pending存在 返回处理中
			if common.CheckTradePendingByTxid(txid) {
				response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
				return
			}
			//都未查到，返回无效的txid
			response.SetResponseBase(constants.RC_INVALID_TXID)
			return
		case constants.TRADE_CURRENCY_ETH:
			//数据库存在 返回成功
			if common.CheckEthHistoryByTxid(txid) {
				return
			}
			//pending存在 返回处理中
			if common.CheckTradePendingByTxid(txid) {
				response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
				return
			}
			//都未查到，返回无效的txid
			response.SetResponseBase(constants.RC_INVALID_TXID)
			return
		case constants.TRADE_CURRENCY_LVT:
			//数据库存在 返回成功
			if common.CheckTXID(txid) {
				return
			}
			//pending存在 返回处理中
			if common.CheckPending(txid) {
				response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
				return
			}
			//都未查到，返回无效的txid
			response.SetResponseBase(constants.RC_INVALID_TXID)
			return
		case constants.TRADE_CURRENCY_LVTC:
			//数据库存在 返回成功
			if common.CheckTXID(txid) {
				return
			}
			//pending存在 返回处理中
			if common.CheckLVTCPending(txid) {
				response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
				return
			}
			//都未查到，返回无效的txid
			response.SetResponseBase(constants.RC_INVALID_TXID)
			return
		}
	}
	response.SetResponseBase(constants.RC_PARAM_ERR)
}
