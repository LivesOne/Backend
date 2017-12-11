package httpHandlers

import (
	"fmt"
	"net/http"
	"servlets/httpcfg"
	"servlets/httputils"
)

// echoMsgHandler implements the "Echo message" interface
type echoMsgHandler struct {
}

func (handler *echoMsgHandler) Method() string {
	return http.MethodPost
}

func (handler *echoMsgHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	msg := request.PostFormValue("param")

	fmt.Println("echoMsgHandler) Handle", msg)

	response := &httputils.ResponseData{
		Base: &httputils.BaseResp{
			RC:  httpCfg.RC_OK,
			Msg: "Success",
		},
		Data: msg,
	}

	// return response

	httputils.FlushJSONData2Client(response, writer)
}
