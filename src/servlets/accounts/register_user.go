package accounts

import (
	"fmt"
	"net/http"
	"servlets/common"
	"servlets/constants"
)

// registerUserHandler implements the "Echo message" interface
type registerUserHandler struct {
}

func (handler *registerUserHandler) Method() string {
	return http.MethodPost
}

func (handler *registerUserHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	// var response *comhttp.ResponseData = comhttp.NewResponseData()

	// response.Base.RC = httpCfg.RC_OK
	// response.Base.Msg = "Success"
	// response.Data = params.Data

	// return response
	// request.ParseForm()

	msg := request.PostFormValue("param")
	// msg := request.FormValue("param")

	fmt.Println("registerUserHandler) Handle", msg)

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
