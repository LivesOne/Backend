package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"utils/logger"
	"utils/config"
	"time"
)

type withdrawQuotaParams struct {
	Uid string `json:"uid"`
}

//type withdrawQuotaRequest struct {
//	Base  *common.BaseInfo  `json:"base"`
//	Param *withdrawQuotaParams `json:"param"`
//}

type withdrawQuotaResponse struct {
	Day    string `json:"day"`
	Month  string `json:"month"`
	Casual string `json:"casual"`
}

type withdrawQuotaHandler struct {
}

func (handler *withdrawQuotaHandler) Method() string {
	return http.MethodPost
}

func (handler *withdrawQuotaHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}

	defer common.FlushJSONData2Client(response, writer)

	requestData := withdrawQuotaParams{}
	common.ParseHttpBodyParams(request, &requestData)

	if requestData.Uid == "" {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if len(requestData.Uid) > 0 {
		uid := utils.Str2Int64(requestData.Uid)
		userWithdrawalQuota := common.GetUserWithdrawalQuotaByUid(uid)
		level := common.GetTransUserLevel(uid)
		limitConfig := config.GetLimitByLevel(level)
		if userWithdrawalQuota == nil {
			userWithdrawalQuota = common.InitUserWithdrawal(uid)
		}

		dayExpend := userWithdrawalQuota.DayExpend
		utils.IsToday(dayExpend, time.Now().Unix())

		if dayExpend > 0 && !utils.IsToday(dayExpend, time.Now().Unix()) {
			if common.ResetDayQuota(uid, utils.FloatStrToLVTint(utils.Int642Str(limitConfig.DailyWithdrawalQuota()))) && time.Now().Day() == 1 {
				common.ResetMonthQuota(uid, utils.FloatStrToLVTint(utils.Int642Str(limitConfig.MonthlyWithdrawalQuota())))
			}
			userWithdrawalQuota = common.GetUserWithdrawalQuotaByUid(uid)
		}

		resData := withdrawQuotaResponse{
			Day:    utils.LVTintToFloatStr(userWithdrawalQuota.Day),
			Month:  utils.LVTintToFloatStr(userWithdrawalQuota.Month),
			Casual: utils.LVTintToFloatStr(userWithdrawalQuota.Casual),
		}
		response.Data = resData
	}
}
