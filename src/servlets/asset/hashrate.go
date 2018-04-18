package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
)

type hashrateParam struct {
	Uid string `json:"uid"`
}

type hashrateRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *hashrateParam   `json:"param"`
}

type hashrateResData struct {
	Hashrate int   `json:"hashrate"`
	Ts       int64 `json:"ts"`
}

// sendVCodeHandler
type hashrateHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *hashrateHandler) Method() string {
	return http.MethodPost
}

func (handler *hashrateHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: 0, // data expire Int 失效时间，单位秒
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := hashrateRequest{} // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &requestData)

	base := requestData.Base

	if base == nil || !base.App.IsValid() {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	intUid := utils.Str2Int64(requestData.Param.Uid)

	if !common.ExistsUID(intUid) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	hr, ts := common.QueryHashRateByUid(intUid)

	response.Data = hashrateResData{
		Hashrate: hr,
		Ts:       ts,
	}

}
