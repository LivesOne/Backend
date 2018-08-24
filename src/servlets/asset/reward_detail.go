package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"gopkg.in/mgo.v2/bson"
)

type tradeParam struct {
	Txid  string `json:"txid"`
	Type  int    `json:"type"`
	Begin int64  `json:"begin"`
	End   int64  `json:"end"`
	Max   int    `json:"max"`
}

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
	Miner     []rewardMiner `json:"miner,omitempty"`
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

	if !common.ExistsUID(intUid) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	re, err := common.QueryLvtcReward(intUid)

	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	yesterday := "0.00000000"
	m := make([]rewardMiner, 0)

	t := re.Lastmodify
	nt := utils.GetTimestamp13()

	//如果时间戳不是昨天，返回0
	if utils.IsToday(t, nt) {
		yesterday = utils.LVTintToFloatStr(re.Yesterday)

		tradeParam := &tradeParam{
			Type: constants.TX_TYPE_REWARD ,
		}

		q := buildTradeQuery(intUid, tradeParam)
		records := common.QueryTrades(q, 1)

		//获取工资明细miner
		m = buildMinerData(records)
	}
	response.Data = rewardDetailResData{
		Total:     utils.LVTintToFloatStr(re.Total),
		Yesterday: yesterday,
		Ts:        t,
		Days:      re.Days,
		Miner:     m,
	}

}

func buildTradeQuery(uid int64, param *tradeParam) bson.M {
	query := bson.M{}

	if len(param.Txid) > 0 {
		query["txid"] = utils.Str2Int64(param.Txid)
	} else {
		//判断时间参数
		ts := []bson.M{}
		if param.Begin > 0 {
			begin := bson.M{
				"txid": bson.M{
					"$gt": utils.TimestampToTxid(param.Begin, 0),
				},
			}
			ts = append(ts, begin)
		}
		if param.End > 0 {
			//end +1 毫秒 为了保证当前毫秒数的记录可以查出来  后22位置0 +1毫秒后的记录不会查出
			end := bson.M{
				"txid": bson.M{
					"$lt": utils.TimestampToTxid(param.End+1, 0),
				},
			}
			ts = append(ts, end)
		}
		if len(ts) > 0 {
			query["$and"] = ts
		}
		//判断查询类型
		//生成不同的查询条件
		switch {
		case param.Type == constants.TX_TYPE_REWARD ||
			param.Type == constants.TX_TYPE_ACTIVITY_REWARD ||
			param.Type == constants.TX_TYPE_PRIVATE_PLACEMENT:
			query["to"] = uid
			query["type"] = param.Type
		case param.Type == constants.TX_TYPE_RECEIVABLES:
			query["to"] = uid
			query["type"] = constants.TX_TYPE_TRANS
		case param.Type == constants.TX_TYPE_TRANS:
			query["from"] = uid
			query["type"] = constants.TX_TYPE_TRANS
		default:
			query["$or"] = []bson.M{
				bson.M{
					"from": uid,
				},
				bson.M{
					"to": uid,
				},
			}
		}
	}

	return query
}

func buildMinerData(records []common.TradeInfo) []rewardMiner {
	m := make([]rewardMiner, 0)

	if records != nil && len(records) > 0 {
		for _, v := range records {
			if len(v.Miner) > 0 {
				for _, item := range v.Miner {
					m = append(m, rewardMiner{
						Sid:   item.Sid,
						Value: utils.LVTintToFloatStr(item.Value),
					})
				}
			}
		}
	}
	return m
}