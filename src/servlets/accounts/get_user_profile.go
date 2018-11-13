package accounts

import (
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)

type profileResponse struct {
	Uid            string           `json:"uid,omitempty"`
	Nickname       string           `json:"nickname,omitempty"`
	Email          string           `json:"email,omitempty"`
	Country        int64            `json:"country,omitempty"`
	Phone          string           `json:"phone,omitempty"`
	HavePayPwd     bool             `json:"have_pay_pwd,omitempty"`
	TransLevel     int              `json:"trans_level,omitempty"`
	BindWx         bool             `json:"bind_wx,omitempty"`
	CreditScore    int64            `json:"credit_score,omitempty"`
	BindTg         bool             `json:"bind_tg,omitempty"`
	WalletAddress  string           `json:"wallet_address,omitempty"`
	AvatarUrl      string           `json:"avatar_url,omitempty"`
	ActiveDays     int64            `json:"active_days,omitempty"`
	Level          int64            `json:"level,omitempty"`
	UpdateTime     int64            `json:"update_time,omitempty"`
	RegisterTime   int64            `json:"register_time,omitempty"`
	HashrateDetial []hashrateDetial `json:"hashrate,omitempty"`
}

type hashrateDetial struct {
	Type  int `json:"type,omitempty"`
	Value int `json:"value,omitempty"`
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

	uid, aesKey, _, tokenErr := rpc.GetTokenInfo(header.TokenHash)
	if tokenErr != microuser.ResCode_OK {
		response.SetResponseBase(rpc.TokenErr2RcErr(tokenErr))
		return
	}

	if !utils.SignValid(aesKey, header.Signature, header.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	intUid := utils.Str2Int64(uid)
	account, err := rpc.GetUserInfo(intUid)
	if err != nil {
		logger.Info("get user profile: read account info failed, uid=", uid)
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}


	wx, _ := rpc.GetUserField(intUid, microuser.UserField_WX)
	tg, _ := rpc.GetUserField(intUid, microuser.UserField_TG)
	updTime, _ := rpc.GetUserField(intUid, microuser.UserField_UPDATE_TIME)
	regTime, _ := rpc.GetUserField(intUid, microuser.UserField_REGISTER_TIME)
	walletAddr, _ := rpc.GetUserField(intUid, microuser.UserField_WALLET_ADDRESS)
	//从缓存中获取用户活跃天数信息
	//提前获取交易等级
	profile := profileResponse{
		Uid:            uid,
		Nickname:       account.Nickname,
		Email:          account.Email,
		Country:        account.Country,
		Phone:          account.Phone,
		HavePayPwd:     len(account.PaymentPassword) > 0,
		TransLevel:     common.GetUserAssetTranslevelByUid(account.Uid),
		BindWx:         len(wx) > 1,
		CreditScore:    account.CreditScore,
		BindTg:         len(tg) > 0,
		WalletAddress:  walletAddr,
		AvatarUrl:      account.AvatarUrl,
		ActiveDays:     account.ActiveDays,
		UpdateTime:     utils.GetTs13(utils.Str2Int64(updTime)),
		RegisterTime:   utils.GetTs13(utils.Str2Int64(regTime)),
		Level:          account.Level,
		HashrateDetial: buildHashrateDetial(account.Uid),
	}

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
