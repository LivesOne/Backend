package accounts

import (
	"errors"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"time"
	"utils"
	"utils/config"
	"utils/db_factory"
	"utils/logger"
	"utils/vcode"
)

type registerUpVcodeParam struct {
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	VCode   string `json:"vcode"`
	PWD     string `json:"pwd"`
	Spkv    int    `json:"spkv"`
}

// registerRequest holds entire request data
type registerUpVcodeRequest struct {
	Base  common.BaseInfo      `json:"base"`
	Param registerUpVcodeParam `json:"param"`
}
type registerUpVcodeHandler struct {
}

func (handler *registerUpVcodeHandler) Method() string {
	return http.MethodPost
}

func (handler *registerUpVcodeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	data := new(registerUpVcodeRequest)
	common.ParseHttpBodyParams(request, data)

	if !checkRegusterUpVcodeRequestParams(header, data) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	hashedPWD, err := recoverHashedPwd(data.Param.PWD, data.Param.Spkv)
	if err != nil {
		logger.Info("register user: decrypt hash pwd error:", err)
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	//validate up vcode

	flag, resErr := vcode.ValidateSmsUpVCode(data.Param.Country, data.Param.Phone, data.Param.VCode)

	if !flag {
		logger.Info("validate up sms code failed")
		response.SetResponseBase(resErr)
		return
	}
	account := new(common.Account)
	account.Country = data.Param.Country
	account.Phone = data.Param.Phone
	account.RegisterTime = time.Now().Unix()
	account.UpdateTime = account.RegisterTime
	account.RegisterType = constants.LOGIN_TYPE_PHONE

	for i := 1; i <= 5; i++ {
		account.UIDString, account.UID = getUid()
		account.LoginPassword = utils.Sha256(hashedPWD + account.UIDString)
		_, err = common.InsertAccountWithPhone(account)
		if err == nil {
			break
		}
		if db_factory.CheckDuplicateByColumn(err, "mobile") {
			response.SetResponseBase(constants.RC_DUP_PHONE)
			return
		} else if db_factory.CheckDuplicateByColumn(err, "PRIMARY") {
			continue
		} else {
			break
		}
	}

}
func checkRegusterUpVcodeRequestParams(header *common.HeaderParams, data *registerUpVcodeRequest) bool {
	if header.Timestamp < 1 {
		logger.Info("register user: no timestamp")
		return false
	}

	if (data.Base.App == nil) || (data.Base.App.IsValid() == false) {
		logger.Info("register user: app info is invalid")
		return false
	}

	if (len(data.Param.PWD) < 1) || (data.Param.Spkv < 1) {
		logger.Info("register user: no password or spkv info")
		return false
	}

	return true
}

func recoverHashedPwd(pwdUpload string, spkv int) (string, error) {

	privKey, err := config.GetPrivateKey(spkv)
	if (err != nil) || (privKey == nil) {
		// logger.Info("register user: load private key failed")
		return "", errors.New("register user: load private key failed")
	}

	// hashPwd, err := utils.RsaDecrypt(string(base64Decode), privKey)
	hashPwd, err := utils.RsaDecrypt(pwdUpload, privKey)
	if err != nil {
		// logger.Info("register user: decrypt pwd error:", err)
		return "", err
	}

	logger.Info("register user: hash pwd:", hashPwd)

	return hashPwd, nil
}
