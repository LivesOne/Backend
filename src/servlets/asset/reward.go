package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"strconv"
)

const  (
	CONV_LVT = 10000*10000
	DAY_1 = 24*60*60*1000
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

	intUid := utils.Str2Int64(requestData.Param.Uid)

	if !common.ExistsUID(intUid) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	re := common.QueryReward(intUid)

	yesterday := formatLVT(re.Yesterday)



	t := re.Lastmodify
	nt := utils.GetTimestamp13()

	if t < (nt - DAY_1){
		yesterday = "0.00"
	}
	//如果时间戳不是昨天，返回0
	response.Data = rewardResData{
		Total:     formatLVT(re.Total),
		Yesterday: yesterday,
	}

}

func formatLVT(lvt int64)string{
	return strconv.FormatFloat((float64(lvt) / CONV_LVT),'f',8,64)
}