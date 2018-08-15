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
	HavePayPwd    bool   `json:"have_pay_pwd"`
	TransLevel    int    `json:"trans_level"`
	BindWx        bool   `json:"bind_wx"`
	CreditScore   int    `json:"credit_score"`
	BindTg        bool   `json:"bind_tg"`
	WalletAddress string `json:"wallet_address"`
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
	if  err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
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
	//提前获取交易等级
	profile := profileResponse{
		HavePayPwd:    (len(account.PaymentPassword) > 0),
		TransLevel:    common.GetUserAssetTranslevelByUid(account.UID),
		BindWx:        bindWx,
		CreditScore:   creditScore,
		BindTg:        bindTg,
		WalletAddress: common.GetUserWalletAddressByUid(account.UID),
	}

	account.ID = 0
	account.UID = 0
	account.LoginPassword = ""
	account.PaymentPassword = ""
	account.From = ""
	account.RegisterType = 0
	//account.WalletAddress = common.GetUserWalletAddressByUid(account.UID)
	profile.Account = *account

	response.Data = profile
}
