package accounts

import (
	"errors"
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"golang.org/x/net/context"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"servlets/vcode"
	"strconv"
	"utils"
	"utils/config"
	"utils/db_factory"
	"utils/logger"
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

	cli := rpc.GetUserCacheClient()
	if cli == nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	req := &microuser.RegUserInfo{
		Pwd:     hashedPWD,
		Country: int32(data.Param.Country),
		Phone:   data.Param.Phone,
		Email:   data.Param.EMail,
		Type:    int32(data.Param.Type),
	}
	resData := new(responseRegister)
	switch data.Param.Type {
	case constants.LOGIN_TYPE_UID:

		//if len(data.Param.VCode) > 0 {
		//	ok, c := vcode.ValidateImgVCode(data.Param.VCodeID, data.Param.VCode)
		//	if ok == false {
		//		response.SetResponseBase(vcode.ConvImgErr(c))
		//		return
		//	}
		//}
		//resp, err := cli.RegisterUser(context.Background(), req)
		//if err != nil {
		//	response.SetResponseBase(constants.RC_SYSTEM_ERR)
		//	return
		//}
		//if resp.Result != microuser.ResCode_OK {
		//	response.SetResponseBase(constants.RC_SYSTEM_ERR)
		//	return
		//}
		//resData.UID = utils.Int642Str(resp.Uid)
		//resData.Regtime = resp.RegTime
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	case constants.LOGIN_TYPE_EMAIL:
		ok, _ := vcode.ValidateMailVCode(data.Param.VCodeID, data.Param.VCode, data.Param.EMail)
		if ok == false {
			response.SetResponseBase(constants.RC_INVALID_VCODE)
			return
		}
		resp, err := cli.RegisterUser(context.Background(), req)


		if err != nil {
			if dupFlag , _ := db_factory.CheckDuplicate(err);!dupFlag{
				response.SetResponseBase(constants.RC_SYSTEM_ERR)
				return
			}
		}
		if resp.Result != microuser.ResCode_OK {
			response.SetResponseBase(constants.RC_DUP_EMAIL)
			return
		}
		resData.UID = utils.Int642Str(resp.Uid)
		resData.Regtime = resp.RegTime
	case constants.LOGIN_TYPE_PHONE:
		ok, _ := vcode.ValidateSmsAndCallVCode(data.Param.Phone, data.Param.Country, data.Param.VCode, 3600, vcode.FLAG_DEF)
		if ok == false {
			response.SetResponseBase(constants.RC_INVALID_VCODE)
			return
		}
		resp, err := cli.RegisterUser(context.Background(), req)
		if err != nil {
			if dupFlag , _ := db_factory.CheckDuplicate(err);!dupFlag{
				response.SetResponseBase(constants.RC_SYSTEM_ERR)
				return
			}
		}
		if resp.Result != microuser.ResCode_OK {
			response.SetResponseBase(constants.RC_DUP_PHONE)
			return
		}
		resData.UID = utils.Int642Str(resp.Uid)
		resData.Regtime = resp.RegTime
	}

	response.Data = resData
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

	uid = common.GenerateUID()
	uid_num, _ = strconv.ParseInt(uid, 10, 64)
	return uid, uid_num
}

// recoverPwd recovery the upload PWD to hash form
// @param: pwdUpload  original upload pwd in http request
func (handler *registerUserHandler) recoverHashedPwd(pwdUpload string, spkv int) (string, error) {

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
