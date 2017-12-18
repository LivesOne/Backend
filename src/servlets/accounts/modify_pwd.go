package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type modifyPwdParam struct {
	Secret string `json:"secret"`
}

type modifyPwdRequest struct {
	// Base  common.BaseInfo `json:"base"`
	Param modifyPwdParam `json:"param"`
}

// modifyPwdHandler
type modifyPwdHandler struct {
	header      *common.HeaderParams // request header param
	requestData *modifyPwdRequest    // request body
}

func (handler *modifyPwdHandler) Method() string {
	return http.MethodPost
}

func (handler *modifyPwdHandler) Handle(request *http.Request, writer http.ResponseWriter) {

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
