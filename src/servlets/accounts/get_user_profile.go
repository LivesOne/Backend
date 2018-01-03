package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils/logger"
)

type profileResponse struct {
	common.Account
	Have_pay_pwd bool `json:"have_pay_pwd"`
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

	uid, errT := token.GetUID(header.TokenHash)
	if (errT != constants.ERR_INT_OK) || (len(uid) != constants.LEN_uid) {
		logger.Info("get user profile: get uid from token cache failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	account, err := common.GetAccountByUID(uid)
	if (err != nil) || (account == nil) {
		logger.Info("get user profile: read account info failed, uid=", uid)
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	profile := profileResponse{
		Have_pay_pwd: (len(account.PaymentPassword) > 0),
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
