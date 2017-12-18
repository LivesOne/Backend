package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type bindPhoneParam struct {
	Action string `json:"action"`
	VCode  string `json:"vcode"`
	Secret string `json:"secret"`
}

type bindPhoneRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param bindPhoneParam `json:"param"`
}

// bindPhoneHandler
type bindPhoneHandler struct {
	header      *common.HeaderParams // request header param
	requestData *bindPhoneRequest    // request body
}

func (handler *bindPhoneHandler) Method() string {
	return http.MethodPost
}

func (handler *bindPhoneHandler) Handle(request *http.Request, writer http.ResponseWriter) {

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
