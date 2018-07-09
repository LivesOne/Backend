package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type ethtransResultParam struct {
	TradeNo string `json:"trade_no"`
}

type ethtransResultRequest struct {
	Base  *common.BaseInfo     `json:"base"`
	Param *ethtransResultParam `json:"param"`
}

// sendVCodeHandler
type ethtransResultHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *ethtransResultHandler) Method() string {
	return http.MethodPost
}

func (handler *ethtransResultHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := ethtransResultRequest{} // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &requestData)

	if requestData.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if len(requestData.Param.TradeNo) > 0 {
		//数据库存在 返回成功
		if common.CheckEthHistory(requestData.Param.TradeNo) {
			return
		}
		////commited存在返回成功
		//if common.CheckCommited(txid) {
		//	return
		//}
		//pending存在 返回处理中
		if common.CheckEthPending(requestData.Param.TradeNo) {
			response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
			return
		}
		//都未查到，返回无效的txid
		response.SetResponseBase(constants.RC_INVALID_TXID)
	} else {
		response.SetResponseBase(constants.RC_PARAM_ERR)
	}

}
