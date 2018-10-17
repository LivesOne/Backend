package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type profileResponse struct {
	common.Account
	HavePayPwd     bool             `json:"have_pay_pwd"`
	TransLevel     int              `json:"trans_level"`
	BindWx         bool             `json:"bind_wx"`
	CreditScore    int              `json:"credit_score"`
	BindTg         bool             `json:"bind_tg"`
	WalletAddress  string           `json:"wallet_address"`
	AvatarUrl      string           `json:"avatar_url"`
	ActiveDays     int              `json:"active_days"`
	HashrateDetial []hashrateDetial `json:"hashrate"`
}

type hashrateDetial struct {
	Type  int `json:"type"`
	Value int `json:"value"`
}

// getProfileHandler
type getProfileHandler struct {
}

func (handler *getProfileHandler) Method() string {
	return http.MethodPost
}

func (handler *getProfileHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	if header.IsValid() == false {
		logger.Info("get user profile: invalid header info")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	uid, aesKey, _, tokenErr := token.GetAll(header.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("get user profile: get uid from token cache failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !utils.SignValid(aesKey, header.Signature, header.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	account, err := common.GetAccountByUID(uid)
	if (err != nil) || (account == nil) {
		logger.Info("get user profile: read account info failed, uid=", uid)
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	bindWx, bindTg, creditScore := common.CheckBindWXByUidAndCreditScore(account.UID, account.Country)
	_, _, _, walletAddress, avatarUrl := common.GetUserExtendByUid(account.UID)
	//从缓存中获取用户活跃天数信息
	activeDays, _ := common.GetCacheUserField(account.UID, common.USER_CACHE_REDIS_FIELD_NAME_ACTIVE_DAYS)
	//提前获取交易等级
	profile := profileResponse{
		HavePayPwd:     (len(account.PaymentPassword) > 0),
		TransLevel:     common.GetUserAssetTranslevelByUid(account.UID),
		BindWx:         bindWx,
		CreditScore:    creditScore,
		BindTg:         bindTg,
		WalletAddress:  walletAddress,
		AvatarUrl:      avatarUrl,
		ActiveDays:     utils.Str2Int(activeDays),
		HashrateDetial: buildHashrateDetial(account.UID),
	}

	account.ID = 0
	account.UID = 0
	account.LoginPassword = ""
	account.PaymentPassword = ""
	account.From = ""
	account.RegisterType = 0
	profile.Account = *account

	response.Data = profile
}

func buildHashrateDetial(uid int64) []hashrateDetial {
	re := make([]hashrateDetial, 0)

	rows := common.QueryHashRateDetailByUid(uid)
	for _, item := range rows {
		t := utils.Str2Int(item["type"])
		v := utils.Str2Int(item["sh"])
		entity := hashrateDetial{
			Type:  t,
			Value: v,
		}
		re = append(re, entity)
	}

	return re
}
