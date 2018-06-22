package asset

import (
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

const (
	MAX_COUNT = 100
	DEF_COUNT = 20
)

type transHistoryParam struct {
	Txid  string `json:"txid"`
	Type  int    `json:"type"`
	Begin int64  `json:"begin"`
	End   int64  `json:"end"`
	Max   int    `json:"max"`
}

type transHistoryRequest struct {
	Param *transHistoryParam `json:"param"`
}

type transHistoryResData struct {
	More    int                  `json:"more"`
	Records []transHistoryRecord `json:"records"`
}

type transHistoryRecord struct {
	Txid  string              `json:"txid"`
	Type  int                 `json:"type"`
	From  string              `json:"from"`
	To    string              `json:"to"`
	Value string              `json:"value"`
	Ts    int64               `json:"ts"`
	Miner []transHistoryMiner `json:"miner,omitempty"`
}

type transHistoryMiner struct {
	Sid   int    `json:"sid"`
	Value string `json:"value"`
}

// sendVCodeHandler
type transHistoryHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *transHistoryHandler) Method() string {
	return http.MethodPost
}

func (handler *transHistoryHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)
	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset transHistory: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset transHistory: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset transHistory: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidString)

	requestData := transHistoryRequest{} // request body
	common.ParseHttpBodyParams(request, &requestData)
	if requestData.Param == nil || !validateType(requestData.Param.Type) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	q := buildQuery(uid, requestData.Param)

	c := DEF_COUNT

	if requestData.Param.Max > MAX_COUNT {
		c = MAX_COUNT
	} else {
		if requestData.Param.Max > 0 {
			c = requestData.Param.Max
		}
	}

	log.Debug(c)
	//query record
	//查新c+1条记录，如果len > c 说明more = 1
	records := common.QueryCommitted(q, c+1)
	response.Data = buildResData(records, c, uid)
}

func buildResData(records []common.DTTXHistory, max int, uid int64) *transHistoryResData {
	data := transHistoryResData{
		More:    0,
		Records: make([]transHistoryRecord, 0),
	}
	if records != nil && len(records) > 0 {
		rcl := len(records)
		if rcl > max {
			data.More = 1
			records = records[:max]
			rcl = max
		}
		for _, v := range records {
			r := transHistoryRecord{
				Txid:  utils.Int642Str(v.Id),
				Type:  convType(v.Type, v.To, uid),
				From:  convUidStr(v.From),
				To:    convUidStr(v.To),
				Value: utils.LVTintToFloatStr(v.Value),
				Ts:    v.Ts,
			}
			if len(v.Miner) > 0 {
				m := make([]transHistoryMiner, 0)
				for _, item := range v.Miner {
					m = append(m, transHistoryMiner{
						Sid:   item.Sid,
						Value: utils.LVTintToFloatStr(item.Value),
					})
				}
				r.Miner = m
			}
			data.Records = append(data.Records, r)
		}
	}
	return &data
}

func validateType(t int) bool {
	if t < constants.TX_TYPE_ALL ||
		t > constants.TX_TYPE_RECEIVABLES {
		return false
	}
	return true
}

func convType(t int, to, uid int64) int {
	if t == constants.TX_TYPE_TRANS {
		if to == uid {
			return constants.TX_TYPE_RECEIVABLES
		}
		return constants.TX_TYPE_TRANS
	}
	return t
}

func convUidStr(uid int64) string {
	if uid == 0 {
		return ""
	}
	return utils.Int642Str(uid)
}

func buildQuery(uid int64, param *transHistoryParam) bson.M {
	query := bson.M{}

	if len(param.Txid) > 0 {
		query["_id"] = utils.Str2Int64(param.Txid)
	} else {
		//判断时间参数
		ts := []bson.M{}
		if param.Begin > 0 {
			begin := bson.M{
				"_id": bson.M{
					"$gt": utils.TimestampToTxid(param.Begin, 0),
				},
			}
			ts = append(ts, begin)
		}
		if param.End > 0 {
			//end +1 毫秒 为了保证当前毫秒数的记录可以查出来  后22位置0 +1毫秒后的记录不会查出
			end := bson.M{
				"_id": bson.M{
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
