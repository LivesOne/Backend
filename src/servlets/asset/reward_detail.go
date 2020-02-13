package asset

import (
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/config"
)

type rewardDetailParam struct {
	Uid string `json:"uid"`
}

type rewardDetailRequest struct {
	Base  *common.BaseInfo   `json:"base"`
	Param *rewardDetailParam `json:"param"`
}

type rewardMiner struct {
	Sid   int    `json:"sid"`
	Value string `json:"value"`
}

type rewardDetailResData struct {
	Total     string        `json:"total"`
	Yesterday string        `json:"yesterday"`
	Ts        int64         `json:"ts"`
	Days      int           `json:"days"`
	Miner     []rewardMiner `json:"miner"`
}

// rewardDetailHandler
type rewardDetailHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *rewardDetailHandler) Method() string {
	return http.MethodPost
}

func (handler *rewardDetailHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := rewardDetailRequest{} // request body
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
	nt := utils.GetTimestamp13()
	yes, tot := getYesterdayAndTotal(t, re.Yesterday, re.Total)
	rData := &rewardDetailResData{
		Total:     tot,
		Yesterday: yes,
		Ts:        t,
		Days:      re.Days,
	}

	//如果时间戳不是昨天，返回0
	if utils.IsToday(t, nt) {
		//rData.Yesterday = utils.LVTintToFloatStr(re.Yesterday)

		q := bson.M{
			"to":       intUid,
			"type":     constants.TRADE_TYPE_MINER,
			"sub_type": constants.TX_SUB_TYPE_WAGE,
		}
		records := common.QueryTrades(q, 1)

		//获取工资明细miner
		rData.Miner = buildMinerData(records)
	}
	response.Data = rData

}

func buildMinerData(records []common.TradeInfo) []rewardMiner {
	m := make([]rewardMiner, 0)
	if records != nil && len(records) > 0 {
		de := int32(config.GetConfig().GetDecimalsByCurrency(constants.TRADE_CURRENCY_LVTC).DBDecimal)
		for _, v := range records {
			if len(v.Miner) > 0 {
				for _, item := range v.Miner {
					m = append(m, rewardMiner{
						Sid:   item.Sid,
						Value: utils.IntToFloatStrByDecimal(item.Value,de,de),
					})
				}
			}
		}
	}
	return m
}
