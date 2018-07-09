package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/config"
	"utils/logger"
	"time"
)

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

	httpHeader := common.ParseHttpHeaderParams(request)

	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset lockList: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset lockList: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}
	uid := utils.Str2Int64(uidString)

	userWithdrawalQuota := common.GetUserWithdrawalQuotaByUid(uid)
	if userWithdrawalQuota == nil {
		userWithdrawalQuota = common.InitUserWithdrawal(uid)
	} else {
		dayExpend := userWithdrawalQuota.DayExpend
		level := common.GetTransUserLevel(uid)
		limitConfig := config.GetLimitByLevel(level)
		if level > userWithdrawalQuota.LastLevel {
			logger.Debug("用户等级提升，uid:", uid, ",原等级：", userWithdrawalQuota.LastLevel, ",现等级：", level)
			oldLimitConfig := config.GetLimitByLevel(userWithdrawalQuota.LastLevel)
			oldLevelDailyQuota := utils.FloatStrToLVTint(utils.Int642Str(oldLimitConfig.DailyWithdrawalQuota()))
			oldLevelMonthlyQuota := utils.FloatStrToLVTint(utils.Int642Str(oldLimitConfig.MonthlyWithdrawalQuota()))
			currentLevelDailyQuota := utils.FloatStrToLVTint(utils.Int642Str(limitConfig.DailyWithdrawalQuota()))
			currentLevelMonthlyQuota := utils.FloatStrToLVTint(utils.Int642Str(limitConfig.MonthlyWithdrawalQuota()))
			balanceDailyQuota := currentLevelDailyQuota - oldLevelDailyQuota
			balanceMonthlyQuota := currentLevelMonthlyQuota - oldLevelMonthlyQuota
			common.ResetDayQuota(uid, balanceDailyQuota+userWithdrawalQuota.Day)
			common.ResetMonthQuota(uid, balanceMonthlyQuota+userWithdrawalQuota.Month)
			common.UpdateLastLevelOfQuota(uid, level)
			userWithdrawalQuota = common.GetUserWithdrawalQuotaByUid(uid)
		} else {
			logger.Debug("用户等级未变化，uid:", uid)
			if dayExpend == 0 || !utils.IsToday(dayExpend, utils.GetTimestamp13()) {
				if userWithdrawalQuota.Day != utils.FloatStrToLVTint(utils.Int642Str(limitConfig.DailyWithdrawalQuota())) {
					logger.Debug("重置日额度，uid:", uid, "，原额度：", userWithdrawalQuota.Day, "，重置额度", utils.FloatStrToLVTint(utils.Int642Str(limitConfig.DailyWithdrawalQuota())))
					common.ResetDayQuota(uid, utils.FloatStrToLVTint(utils.Int642Str(limitConfig.DailyWithdrawalQuota())))
				}

				lastExpendDate := utils.Timestamp13ToDate(userWithdrawalQuota.DayExpend)
				if lastExpendDate.Year() < time.Now().Year() || (lastExpendDate.Year() == time.Now().Year() && lastExpendDate.Month() < time.Now().Month()) {
					if userWithdrawalQuota.Month != utils.FloatStrToLVTint(utils.Int642Str(limitConfig.MonthlyWithdrawalQuota())) {
						logger.Debug("重置月额度，uid:", uid, "，原额度：", userWithdrawalQuota.Day, "，重置额度", utils.FloatStrToLVTint(utils.Int642Str(limitConfig.DailyWithdrawalQuota())))
						common.ResetMonthQuota(uid, utils.FloatStrToLVTint(utils.Int642Str(limitConfig.MonthlyWithdrawalQuota())))
					}
				}

				userWithdrawalQuota = common.GetUserWithdrawalQuotaByUid(uid)
			}
		}
	}

	resData := withdrawQuotaResponse{
		Day:    utils.LVTintToFloatStr(userWithdrawalQuota.Day),
		Month:  utils.LVTintToFloatStr(userWithdrawalQuota.Month),
		Casual: utils.LVTintToFloatStr(userWithdrawalQuota.Casual),
	}
	response.Data = resData
}
