package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)




type withdrawCardListResData struct {
	Cards []withdrawCardListRecord `json:"cards"`
}

type withdrawCardListRecord struct {
	Id string `json:"id"`
	Password string `json:"password"`
	Quota string `json:"quota"`
	Expire int64 `json:"expire"`
	Cost string `json:"cost"`
	GetTime int64 `json:"get_time"`
	UseTime int64 `json:"use_time"`
	Status int `json:"status"`
}

// sendVCodeHandler
type withdrawCardListHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *withdrawCardListHandler) Method() string {
	return http.MethodPost
}

func (handler *withdrawCardListHandler) Handle(request *http.Request, writer http.ResponseWriter) {
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
		log.Info("asset withdrawCardList: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset withdrawCardList: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset withdrawCardList: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidString)

	dbRecord := common.GetUserWithdrawCardByUid(uid)
	if len(dbRecord) > 0 {
		response.Data = &withdrawCardListResData{
			Cards: convRowToWithdrawCardListRecord(dbRecord),
		}
	}

}

func convRowToWithdrawCardListRecord(rows []map[string]string)[]withdrawCardListRecord{
	re := make([]withdrawCardListRecord,0)

	for _,item := range rows {
		quota :=  utils.LVTintToFloatStr(utils.Str2Int64(item["quota"]))
		cost := utils.LVTintToFloatStr(utils.Str2Int64(item["cost"]))
		entity := withdrawCardListRecord{
			Id:       item["id"],
			Password: item["password"],
			Quota:    quota,
			Expire:   utils.Str2Int64(item["expire_time"]),
			Cost:     cost,
			GetTime:  utils.Str2Int64(item["get_time"]),
			UseTime:  utils.Str2Int64(item["use_time"]),
			Status:   utils.Str2Int(item["status"]),
		}
		re = append(re,entity)
	}
	return re
}