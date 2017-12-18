package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type bindEMailParam struct {
	Action string `json:"action"`
	VCode  string `json:"vcode"`
	Secret string `json:"secret"`
}

type bindEMailRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param bindEMailParam `json:"param"`
}

// bindEMailHandler
type bindEMailHandler struct {
	header      *common.HeaderParams // request header param
	requestData *bindEMailRequest    // request body
}

func (handler *bindEMailHandler) Method() string {
	return http.MethodPost
}

func (handler *bindEMailHandler) Handle(request *http.Request, writer http.ResponseWriter) {

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
