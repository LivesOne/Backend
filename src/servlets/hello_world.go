package httpHandlers

import (
	"fmt"
	"net/http"
	"servlets/httpcfg"
	"servlets/httputils"
)

// helloWorldHandler implements the "Echo message" interface
type helloWorldHandler struct {
}

func (handler *helloWorldHandler) Method() string {
	return http.MethodGet
}

func (handler *helloWorldHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	msg := request.FormValue("param")

	headerparam := httputils.ParseHttpHeaderParams(request)

	fmt.Println("helloWorldHandler, msg&HeaderParam:", headerparam, msg)

	response := &httputils.ResponseData{
		Base: &httputils.BaseResp{
			RC:  httpCfg.RC_OK,
			Msg: "Success",
		},
		Data: headerparam,
	}

	// return response

	httputils.FlushJSONData2Client(response, writer)
}
