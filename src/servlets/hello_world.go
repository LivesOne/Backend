package servlets

import (
	"fmt"
	"net/http"
	"servlets/common"
	"servlets/constants"
)

// helloWorldHandler implements the "Echo message" interface
type helloWorldHandler struct {
}

func (handler *helloWorldHandler) Method() string {
	return http.MethodGet
}

func (handler *helloWorldHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	msg := request.FormValue("param") // actually, this is URL param
	headerparam := common.ParseHttpHeaderParams(request)

	fmt.Println("helloWorldHandler, msg&HeaderParam:", msg, headerparam)

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: msg,
	}

	common.FlushJSONData2Client(response, writer)
}
