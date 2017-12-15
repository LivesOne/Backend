package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type logoutRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param string          `json:"param"`
}

// logoutHandler implements the "Echo message" interface
type logoutHandler struct {
	header    *common.HeaderParams // request header param
	logouData *logoutRequest       // request login data
}

func (handler *logoutHandler) Method() string {
	return http.MethodPost
}

func (handler *logoutHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.logouData)
}
