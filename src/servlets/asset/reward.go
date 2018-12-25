package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/config"
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
	Ts        int64  `json:"ts"`
	Days      int    `json:"days"`
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
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := rewardRequest{} // request body
	//header := common.ParseHttpHeaderParams(request)
	if !common.ParseHttpBodyParams(request, &requestData) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	base := requestData.Base

	if base == nil || !base.App.IsValid() {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	intUid := utils.Str2Int64(requestData.Param.Uid)

	if !rpc.UserExists(intUid) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	re, err := common.QueryLvtcReward(intUid)

	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}


	t := re.Lastmodify

	yes,tot := getYesterdayAndTotal(t,re.Yesterday,re.Total)
	rpc.ActiveUser(intUid)
	response.Data = rewardResData{
		Total:     tot,
		Yesterday: yes,
		Ts:        t,
		Days:      re.Days,
	}

}


func getYesterdayAndTotal(ts,yes,tot int64)(string,string){
	nt := utils.GetTimestamp13()
	de := int32(config.GetConfig().GetDecimalsByCurrency(constants.TRADE_CURRENCY_LVTC).DBDecimal)

	total := utils.IntToFloatStrByDecimal(tot,de,de)

	if utils.IsToday(ts, nt) {
		return utils.IntToFloatStrByDecimal(yes,de,de),total
	}
	return utils.IntToFloatStrByDecimal(0,de,de),total
}