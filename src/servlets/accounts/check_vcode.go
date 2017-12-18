package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type checkVCodeParam struct {
	Type    int    `json:"type"`
	Action  string `json:"action"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	VCode   string `json:"vcode"`
	Keep    int    `json:"keep"`
}

type checkVCodeRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param checkVCodeParam `json:"param"`
}

// checkVCodeHandler
type checkVCodeHandler struct {
	header      *common.HeaderParams // request header param
	requestData *checkVCodeRequest   // request body
}

func (handler *checkVCodeHandler) Method() string {
	return http.MethodPost
}

func (handler *checkVCodeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

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
