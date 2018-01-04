package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
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

}
