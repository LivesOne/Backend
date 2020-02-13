package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)

type ethtransHistoryParam struct {
	Txid  string `json:"txid"`
	Type  int    `json:"type"`
	Begin int64  `json:"begin"`
	End   int64  `json:"end"`
	Max   int    `json:"max"`
}

type ethtransHistoryRequest struct {
	Param *ethtransHistoryParam `json:"param"`
}

type ethtransHistoryResData struct {
	More    int                     `json:"more"`
	Records []ethtransHistoryRecord `json:"records"`
}

type ethtransHistoryRecord struct {
	Txid  string `json:"txid"`
	Type  int    `json:"type"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
	Ts    int64  `json:"ts"`
}

// sendVCodeHandler
type ethtransHistoryHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *ethtransHistoryHandler) Method() string {
	return http.MethodPost
}

func (handler *ethtransHistoryHandler) Handle(request *http.Request, writer http.ResponseWriter) {
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
		log.Info("asset ethtransHistory: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset ethtransHistory: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset ethtransHistory: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidString)

	requestData := new(ethtransHistoryRequest) // request body

	if !common.ParseHttpBodyParams(request, requestData) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	p := requestData.Param

	if p == nil || !validateEthType(p.Type) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	c := DEF_COUNT

	if p.Max > MAX_COUNT {
		c = MAX_COUNT
	} else {
		if p.Max > 0 {
			c = requestData.Param.Max
		}
	}

	//query record
	//查新c+1条记录，如果len > c 说明more = 1

	dbRecord := common.QueryEthTxHistory(uid, p.Txid, p.Type, p.Begin, p.End, c+1)

	resData := &ethtransHistoryResData{
		More:    0,
		Records: nil,
	}
	recordLen := len(dbRecord)
	if recordLen > 0 {
		var resListRecords []ethtransHistoryRecord
		if recordLen > c {
			resData.More = 1
			resListRecords = convRowToTxHistoryRecord(dbRecord[:c])
		} else {
			resListRecords = convRowToTxHistoryRecord(dbRecord)
		}
		resData.Records = resListRecords
	}
	response.Data = resData

}

func convRowToTxHistoryRecord(rows []map[string]string) []ethtransHistoryRecord {
	re := make([]ethtransHistoryRecord, 0)

	for _, item := range rows {
		value := utils.Str2Int64(item["value"])
		entity := ethtransHistoryRecord{
			Txid:  item["txid"],
			Type:  utils.Str2Int(item["type"]),
			From:  item["from"],
			To:    item["to"],
			Value: utils.LVTintToFloatStr(value),
			Ts:    utils.Str2Int64(item["ts"]),
		}
		re = append(re, entity)
	}
	return re
}

func validateEthType(t int) bool {
	if t == constants.TX_TYPE_ALL ||
		(t >= constants.TX_TYPE_RECHANGE &&
			t <= constants.TX_TYPE_BUY_COIN_CARD) {
		return true
	}
	return false
}
