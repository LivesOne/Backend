package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

// getProfileHandler
type getProfileHandler struct {
	header *common.HeaderParams // request header param
	// requestData *logoutRequest       // request body
}

func (handler *getProfileHandler) Method() string {
	return http.MethodPost
}

func (handler *getProfileHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	// common.ParseHttpBodyParams(request, &handler.requestData)
}
