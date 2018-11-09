package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)

type lvtcTransHistoryMiner struct {
	Sid   int    `json:"sid"`
	Value string `json:"value"`
}

// sendVCodeHandler
type lvtcTransHistoryHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *lvtcTransHistoryHandler) Method() string {
	return http.MethodPost
}

func (handler *lvtcTransHistoryHandler) Handle(request *http.Request, writer http.ResponseWriter) {
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
		log.Info("asset lvtcTransHistory: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset lvtcTransHistory: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset lvtcTransHistory: get aeskey from cache error:", len(aesKey))
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
	records := common.QueryLVTCCommitted(q, c+1)
	response.Data = buildResData(records, c, uid)
}
