package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type setTxPwdParam struct {
	Action string `json:"action"`
	Type   int    `json:"type"`
	VCode  string `json:"vcode"`
	PWD    string `json:"pwd"`
}

type setTxPwdRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param setTxPwdParam `json:"param"`
}

// setTxPwdHandler
type setTxPwdHandler struct {
	header      *common.HeaderParams // request header param
	requestData *setTxPwdRequest     // request body
}

func (handler *setTxPwdHandler) Method() string {
	return http.MethodPost
}

func (handler *setTxPwdHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.requestData)
}
