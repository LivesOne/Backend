package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"strconv"
)


type rewardParam struct {
	Uid string `json:"uid"`
}

type rewardRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *rewardParam     `json:"param"`
}

type rewardResData struct {
	Total     string `json:"total"`
	Yesterday string `json:"yesterday"`
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

	re := common.QueryReward(intUid)

	yesterday := formatLVT(re.Yesterday)



	t := re.Lastmodify
	nt := utils.GetTimestamp13()

	if utils.IsNextDay(t,nt){
		yesterday = "0.00"
	}
	//如果时间戳不是昨天，返回0
	response.Data = rewardResData{
		Total:     utils.LVTintToFloatStr(re.Total),
		Yesterday: yesterday,
	}

}

