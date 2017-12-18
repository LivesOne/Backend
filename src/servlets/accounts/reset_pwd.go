package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type resetPwdParam struct {
	Action  string `json:"action"`
	Type    int    `json:"type"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	VCode   string `json:"vcode"`
	PWD     string `json:"pwd"`
	Spkv    int    `json:"spkv"`
}

type resetPwdRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param resetPwdParam   `json:"param"`
}

// resetPwdHandler
type resetPwdHandler struct {
	header      *common.HeaderParams // request header param
	requestData *resetPwdRequest     // request body
}

func (handler *resetPwdHandler) Method() string {
	return http.MethodPost
}

func (handler *resetPwdHandler) Handle(request *http.Request, writer http.ResponseWriter) {

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
