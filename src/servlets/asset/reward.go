package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type rewardParam struct {

}

type rewardRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *rewardParam  `json:"param"`
}

// sendVCodeHandler
type rewardHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *rewardHandler) Method() string {
	return http.MethodPost
}

func (handler *rewardHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: 0, // data expire Int 失效时间，单位秒
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := rewardRequest{} // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &requestData)


}
