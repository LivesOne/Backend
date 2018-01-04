package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
)

type transCommitParam struct {
	Txid string `json:"txid"`
}

type transCommitRequest struct {
	Base  *common.BaseInfo   `json:"base"`
	Param *transCommitParam `json:"param"`
}

// sendVCodeHandler
type transCommitHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *transCommitHandler) Method() string {
	return http.MethodPost
}

func (handler *transCommitHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: 0, // data expire Int 失效时间，单位秒
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := transCommitRequest{} // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &requestData)

	//TODO 校验token对应的uid 和panding 中的from uid 是否一致
	txid := utils.Str2Int64(requestData.Param.Txid)
	pending := common.FindPending(txid)

	pending.Status = constants.TX_STATUS_COMMIT

	perPending := common.FindAndModify(txid,pending)

	if perPending.Id == txid && perPending.Status == constants.TX_STATUS_DEFAULT {
		//TODO  判断to是否存在


		if !common.TransAccountLvt(perPending.From,perPending.To,perPending.Value) {

		}
	}

}
