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
	"utils/logger"
	"utils/db_factory"
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

	if checkRequestParams(header,&data) == false {
		setResponseBase(response,constants.RC_PARAM_ERR)
		return
	}

	// fmt.Println("registerUserHandler) Handle", msg)
	// hashPwd := utils.RsaDecrypt(handler.registerData.Param.PWD, config.GetConfig().PrivKey)

	account, err := getAccount(&data)
	if err != nil {
		// logger.Info("------------- get account error\n")
		setResponseBase(response,constants.RC_INVALID_PUB_KEY)
		return
	}
	logger.Info("------------- get account success\n", utils.ToJSONIndent(account))

	switch data.Param.Type {
	case constants.LOGIN_TYPE_UID:
		insertAndCheckUid(account)
	case constants.LOGIN_TYPE_EMAIL:
		f,_ := vcode.ValidateMailVCode(data.Param.VCodeID,data.Param.VCode,data.Param.EMail)
		if f {
			_, err = common.InsertAccountWithEmail(account)
			if err!=nil{
				if  db_factory.CheckDuplicateByColumn(err,"email"){
					setResponseBase(response,constants.RC_DUP_EMAIL)
				}else if  db_factory.CheckDuplicateByColumn(err,"uid"){
					account.UIDString,account.UID = getUid()
					e := insertAndCheckUid(account)
					if e != nil {
						if db_factory.CheckDuplicateByColumn(err,"email"){
							setResponseBase(response,constants.RC_DUP_EMAIL)
						}else{
							setResponseBase(response,constants.RC_SYSTEM_ERR)
						}
					}
				}
				return
			}
		}else{
			setResponseBase(response,constants.RC_INVALID_VCODE)
			return
		}
	case constants.LOGIN_TYPE_PHONE:
		f,_ := vcode.ValidateSmsAndCallVCode(data.Param.Phone,data.Param.Country,data.Param.VCode,3600,vcode.FLAG_DEF)
		if f {
			_, err = common.InsertAccountWithPhone(account)
			if err!=nil{
				if  db_factory.CheckDuplicateByColumn(err,"phone"){
					setResponseBase(response,constants.RC_DUP_PHONE)
				}else if  db_factory.CheckDuplicateByColumn(err,"uid"){
					account.UIDString,account.UID = getUid()
					e := insertAndCheckUid(account)
					if e != nil {
						if db_factory.CheckDuplicateByColumn(err,"phone"){
							setResponseBase(response,constants.RC_DUP_PHONE)
						}else{
							setResponseBase(response,constants.RC_SYSTEM_ERR)
						}
					}
				}
				return
			}
		}else{
			setResponseBase(response,constants.RC_INVALID_VCODE)
			return
		}
	}

	if err != nil {
		setResponseBase(response,constants.RC_SYSTEM_ERR)
		return
	}

	response.Data = &responseRegister{
		UID:     account.UIDString,
		Regtime: account.RegisterTime,
	}
}

func  setResponseBase(resData *common.ResponseData,error constants.Error) {
	resData.Base.RC = error.Rc
	resData.Base.Msg = error.Msg
	logger.Info(error.Msg)
}

func checkRequestParams(header *common.HeaderParams ,data *registerRequest) bool {
	if header.Timestamp < 1 {
		return false
	}

	if (data.Base.App == nil) || (data.Base.App.IsValid() == false) {
		return false
	}

	if (data.Param.Type < constants.LOGIN_TYPE_UID) || (data.Param.Type > constants.LOGIN_TYPE_PHONE) {
		return false
	}

	if data.Param.Type == constants.LOGIN_TYPE_EMAIL && len(data.Param.EMail) < 1 {
		return false
	}

	if data.Param.Type == constants.LOGIN_TYPE_PHONE && (data.Param.Country == 0 || len(data.Param.Phone) < 1) {
		return false
	}

	if (len(data.Param.PWD) < 1) || (data.Param.Spkv < 1) {
		return false
	}

	return true
}

func getUid()(string,int64){
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
	return uid,uid_num
}

func insertAndCheckUid(account *common.Account)error{
	var err error
	for {
		_, err = common.InsertAccount(account)
		if err == nil {
			break
		}else{
			if db_factory.CheckDuplicateByColumn(err,"uid"){
				account.UIDString,account.UID = getUid()
			}else{
				break
			}
		}
	}
	return err
}

func getAccount(data *registerRequest) (*common.Account, error) {
	var account common.Account

	recoverPWD, err := recoverPwd(data)
	if err != nil {
		return nil, err
	}

	account.UIDString,account.UID = getUid()

	account.Email = data.Param.EMail
	account.Country = data.Param.Country
	account.Phone = data.Param.Phone

	account.LoginPassword = utils.Sha256(recoverPWD + account.UIDString)
	account.RegisterTime = time.Now().Unix()
	account.UpdateTime = account.RegisterTime
	account.RegisterType = data.Param.Type

	return &account, nil
}

// recoverPwd recovery the upload PWD to hash form
func recoverPwd(data *registerRequest) (string, error) {

	privKey := config.GetPrivateKey()
	if privKey == nil {
		// fmt.Println("11111111111111:")
		return "", errors.New("load private key failed")
	}

	// fmt.Println("2222222222222222:ggggggggggggggg")
	// hashPwd, err := utils.RsaDecrypt(string(base64Decode), privKey)
	hashPwd, err := utils.RsaDecrypt(data.Param.PWD, privKey)
	if err != nil {
		// fmt.Println("2222222222222222:", err)
		logger.Info("decrypt pwd error:", err)
		return "", err
	}

	logger.Info("----------hash pwd:", hashPwd)
	return string(hashPwd), nil
}
