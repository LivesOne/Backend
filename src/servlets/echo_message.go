package servlets

import (
	"fmt"
	"net/http"
	"servlets/common"
	"servlets/constants"
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

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK,
			Msg: "Success",
		},
		Data: msg,
	}

	// return response

	common.FlushJSONData2Client(response, writer)
}
