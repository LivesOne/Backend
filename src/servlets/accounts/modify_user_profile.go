package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

// modifyUserProfileHandler
type modifyUserProfileHandler struct {
	header *common.HeaderParams // request header param
	// requestData *logoutRequest       // request body
}

func (handler *modifyUserProfileHandler) Method() string {
	return http.MethodPost
}

func (handler *modifyUserProfileHandler) Handle(request *http.Request, writer http.ResponseWriter) {

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
