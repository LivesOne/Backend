package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)




type withdrawCardUseListResData struct {
	Cards []withdrawCardUseListRecord `json:"records"`
}

type withdrawCardUseListRecord struct {
	Id string `json:"id"`
	TradeNo string `json:"trade_no"`
	Txid string `json:"txid"`
	Type int `json:"type"`
	Cost string `json:"cost"`
	Quota string `json:"quota"`
	UseTime int64 `json:"use_time"`
}

// sendVCodeHandler
type withdrawCardUseListHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *withdrawCardUseListHandler) Method() string {
	return http.MethodPost
}

func (handler *withdrawCardUseListHandler) Handle(request *http.Request, writer http.ResponseWriter) {
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
		log.Info("asset withdrawCardUseList: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset withdrawCardUseList: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset withdrawCardUseList: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidString)

	dbRecord := common.GetUserWithdrawCardUseByUid(uid)
	if len(dbRecord) > 0 {
		response.Data = &withdrawCardUseListResData{
			Cards: convRowTowithdrawCardUseListRecord(dbRecord),
		}
	}

}

func convRowTowithdrawCardUseListRecord(rows []map[string]string)[]withdrawCardUseListRecord{
	re := make([]withdrawCardUseListRecord,0)

	for _,item := range rows {
		quota :=  utils.LVTintToFloatStr(utils.Str2Int64(item["quota"]))
		cost := utils.LVTintToFloatStr(utils.Str2Int64(item["cost"]))
		entity := withdrawCardUseListRecord{
			Id:       item["id"],
			Txid: 	  item["txid"],
			Type: 	  utils.Str2Int(item["type"]),
			Quota:    quota,
			TradeNo:  item["trade_no"],
			Cost:     cost,
			UseTime:  utils.Str2Int64(item["create_time"]),
		}
		re = append(re,entity)
	}
	return re
}