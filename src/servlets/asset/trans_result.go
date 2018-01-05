package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
)

type transResultParam struct {
	Txid string `json:"txid"`
}

type transResultRequest struct {
	Base  *common.BaseInfo   `json:"base"`
	Param *transResultParam `json:"param"`
}

// sendVCodeHandler
type transResultHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *transResultHandler) Method() string {
	return http.MethodPost
}

func (handler *transResultHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: 0, // data expire Int 失效时间，单位秒
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := transResultRequest{} // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &requestData)


	if requestData.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if len(requestData.Param.Txid) >0 {
		txid := utils.Str2Int64(requestData.Param.Txid)
		//数据库存在 返回成功
		if common.CheckTXID(txid){
			return
		}
		//commited存在返回成功
		if common.CheckCommited(txid) {
			return
		}
		//pending存在 返回处理中
		if common.CheckPending(txid) {
			response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
			return
		}
		//都未查到，返回无效的txid
		response.SetResponseBase(constants.RC_INVALID_TXID)
	}else {
		response.SetResponseBase(constants.RC_PARAM_ERR)
	}





}
