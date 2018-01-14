package accounts

import (
	"errors"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strconv"
	"time"
	"utils"
	"utils/config"
	"utils/db_factory"
	"utils/logger"
	"utils/vcode"
)

// registerParam holds the request "param" field
type registerParam struct {
	Type    int    `json:"type"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	VCodeID string `json:"vcode_id"`
	VCode   string `json:"vcode"`
	PWD     string `json:"pwd"`
	Spkv    int    `json:"spkv"`
}

// registerRequest holds entire request data
type registerRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param registerParam   `json:"param"`
}

// responseData holds response "data" field
type responseRegister struct {
	UID     string `json:"uid"`
	Regtime int64  `json:"regtime"`
}

// registerUserHandler implements the "Echo message" interface
type registerUserHandler struct {
	// http request, header params
	//header *common.HeaderParams
	// http request, body params
	//registerData *registerRequest

	// http response data to client
	//response *common.ResponseData

	// hashedPWD upload by client
	// hashedPWD string
}

func (handler *registerUserHandler) Method() string {
	return http.MethodPost
}

func (handler *registerUserHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	data := registerRequest{}
	common.ParseHttpBodyParams(request, &data)

	if checkRequestParams(header, &data) == false {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	hashedPWD, err := handler.recoverHashedPwd(data.Param.PWD, data.Param.Spkv)
	if err != nil {
		logger.Info("register user: decrypt hash pwd error:", err)
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	// fmt.Println("registerUserHandler) Handle", msg)
	// hashPwd := utils.RsaDecrypt(handler.registerData.Param.PWD, config.GetConfig().PrivKey)

	account, err := getAccount(&data)
	if err != nil {
		// logger.Info("------------- get account error\n")
		response.SetResponseBase(constants.RC_INVALID_PUB_KEY)
		return
	}
	logger.Info("register user:  get account success\n", utils.ToJSONIndent(account))

	switch data.Param.Type {
	case constants.LOGIN_TYPE_UID:
		insertAndCheckUid(account, hashedPWD)
	case constants.LOGIN_TYPE_EMAIL:
		ok, _ := vcode.ValidateMailVCode(data.Param.VCodeID, data.Param.VCode, data.Param.EMail)
		if ok == false {
			response.SetResponseBase(constants.RC_INVALID_VCODE)
			return
		}

		for {
			account.UIDString, account.UID = getUid()
			account.LoginPassword = utils.Sha256(hashedPWD + account.UIDString)
			_, err = common.InsertAccountWithEmail(account)
			if err == nil {
				break
			}
			if db_factory.CheckDuplicateByColumn(err, "email") {
				response.SetResponseBase(constants.RC_DUP_EMAIL)
				return
			} else if db_factory.CheckDuplicateByColumn(err, "uid") {
				continue
			}
		}
	case constants.LOGIN_TYPE_PHONE:
		ok, _ := vcode.ValidateSmsAndCallVCode(data.Param.Phone, data.Param.Country, data.Param.VCode, 3600, vcode.FLAG_DEF)
		if ok == false {
			response.SetResponseBase(constants.RC_INVALID_VCODE)
			return
		}

		for {
			account.UIDString, account.UID = getUid()
			account.LoginPassword = utils.Sha256(hashedPWD + account.UIDString)
			_, err = common.InsertAccountWithPhone(account)
			if err == nil {
				break
			}
			if db_factory.CheckDuplicateByColumn(err, "mobile") {
				response.SetResponseBase(constants.RC_DUP_PHONE)
				return
			} else if db_factory.CheckDuplicateByColumn(err, "uid") {
				continue
			}
		}
	}

	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	response.Data = &responseRegister{
		UID:     account.UIDString,
		Regtime: account.RegisterTime,
	}
}

func checkRequestParams(header *common.HeaderParams, data *registerRequest) bool {
	if header.Timestamp < 1 {
		logger.Info("register user: no timestamp")
		return false
	}

	if (data.Base.App == nil) || (data.Base.App.IsValid() == false) {
		logger.Info("register user: app info is invalid")
		return false
	}

	if (data.Param.Type < constants.LOGIN_TYPE_UID) || (data.Param.Type > constants.LOGIN_TYPE_PHONE) {
		logger.Info("register user: register type invalid")
		return false
	}

	if data.Param.Type == constants.LOGIN_TYPE_EMAIL && (utils.IsValidEmailAddr(data.Param.EMail) == false) {
		logger.Info("register user: email info invalid")
		return false
	}

	if data.Param.Type == constants.LOGIN_TYPE_PHONE && (data.Param.Country == 0 || len(data.Param.Phone) < 1) {
		logger.Info("register user: phone info invalid")
		return false
	}

	if (len(data.Param.PWD) < 1) || (data.Param.Spkv < 1) {
		logger.Info("register user: no password or spkv info")
		return false
	}

	return true
}

func getUid() (string, int64) {
	var uid string
	var uid_num int64

	//for {
	//	uid = common.GenerateUID()
	//	uid_num, _ = strconv.ParseInt(uid, 10, 64)
	//
	//	if common.ExistsUID(uid_num) {
	//		continue
	//	} else {
	//		break
	//	}
	//}
	uid = common.GenerateUID()
	uid_num, _ = strconv.ParseInt(uid, 10, 64)
	return uid, uid_num
}

func insertAndCheckUid(account *common.Account, hashedPWD string) error {
	var err error
	for {
		account.UIDString, account.UID = getUid()
		account.LoginPassword = utils.Sha256(hashedPWD + account.UIDString)
		_, err = common.InsertAccount(account)
		if err == nil {
			break
		}
		if db_factory.CheckDuplicateByColumn(err, "uid") {
			continue
		}
	}
	return err
}

func getAccount(data *registerRequest) (*common.Account, error) {
	var account common.Account

	// recoverPWD, err := recoverPwd(data)
	// if err != nil {
	// 	return nil, err
	// }

	// account.UIDString, account.UID = getUid()  // delay this before insert to DB

	account.Email = data.Param.EMail
	account.Country = data.Param.Country
	account.Phone = data.Param.Phone

	// account.LoginPassword = utils.Sha256(recoverPWD + account.UIDString)
	account.RegisterTime = time.Now().Unix()
	account.UpdateTime = account.RegisterTime
	account.RegisterType = data.Param.Type

	return &account, nil
}

// recoverPwd recovery the upload PWD to hash form
// @param: pwdUpload  original upload pwd in http request
func (handler *registerUserHandler) recoverHashedPwd(pwdUpload string, spkv int) (string, error) {

	privKey, err := config.GetPrivateKey(spkv)
	if (err != nil) || (privKey == nil) {
		// response.SetResponseBase(constants.RC_PARAM_ERR)
		// logger.Info("register user: load private key failed")
		return "", errors.New("register user: load private key failed")
	}

	// fmt.Println("2222222222222222:ggggggggggggggg")
	// hashPwd, err := utils.RsaDecrypt(string(base64Decode), privKey)
	hashPwd, err := utils.RsaDecrypt(pwdUpload, privKey)
	if err != nil {
		// fmt.Println("2222222222222222:", err)
		// logger.Info("register user: decrypt pwd error:", err)
		return "", err
	}

	logger.Info("register user: hash pwd:", hashPwd)

	return hashPwd, nil
}
