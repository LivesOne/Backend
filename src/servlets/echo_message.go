package servlets

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils/logger"
)

// echoMsgHandler implements the "Echo message" interface
type echoMsgHandler struct {
}

func (handler *echoMsgHandler) Method() string {
	return http.MethodPost
}

func (handler *echoMsgHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	type reqParam struct {
		Param string `json:"param"`
	}

	var params reqParam
	// msg := request.PostFormValue("param")
	common.ParseHttpBodyParams(request, &params)

	logger.Info("echo msg ---> received http body: ", params.Param)

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK,
			Msg: "Success",
		},
		Data: params.Param,
	}

	// return response

	common.FlushJSONData2Client(response, writer)
}
