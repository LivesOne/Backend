package asset

import (
	"net/http"
	"utils/logger"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
)

type withdrawListParams struct {
	AuthType  int    `json:"auth_type"`
	QuotaType int    `json:"quota_type"`
	VcodeType int    `json:"vcode_type"`
	VcodeId   string `json:"vcode_id"`
	Vcode     string `json:"vcode"`
	Secret    string `json:"secret"`
}

type withdrawListResponseData struct {
	Records []withdrawListResponse `json:"records"`
}

type withdrawListResponse struct {
	Id         string  `json:"id"`
	TradeNo    string `json:"trade_no"`
	AssetType  string `json:"asset_type"`
	Address    string `json:"address"`
	Value      string `json:"value"`
	Fee        string `json:"fee"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
	Status     int    `json:"status"`
}

type withdrawListHandler struct {
}

func (handler *withdrawListHandler) Method() string {
	return http.MethodPost
}

func (handler *withdrawListHandler) Handle(request *http.Request, writer http.ResponseWriter) {
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

	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset lockList: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset lockList: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}
	uid := utils.Str2Int64(uidString)

	userWithdrawalRequestArray := common.QueryWithdrawalList(uid)
	withdrawListResponseArray := make([]withdrawListResponse, 0)
	for _, userWithdrawalRequest := range userWithdrawalRequestArray {
		withdrawListResponseArray = append(withdrawListResponseArray, withdrawListResponse{
			Id:         utils.Int642Str(userWithdrawalRequest.Id),
			TradeNo:    userWithdrawalRequest.TradeNo,
			AssetType:  "LVT",
			Address:    userWithdrawalRequest.Address,
			Value:      utils.LVTintToFloatStr(userWithdrawalRequest.Value),
			Fee:        utils.LVTintToFloatStr(userWithdrawalRequest.Free),
			CreateTime: userWithdrawalRequest.CreateTime,
			UpdateTime: userWithdrawalRequest.UpdateTime,
			Status:     userWithdrawalRequest.Status,
		})
	}
	response.Data = &withdrawListResponseData{
		Records: withdrawListResponseArray,
	}
}
