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
	"utils"
	"utils/config"
	"utils/logger"
)

type registerUpVcodeParam struct {
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	VCode   string `json:"vcode"`
	PWD     string `json:"pwd"`
	Spkv    int    `json:"spkv"`
}

type registerUpVcodeResData struct {
	Uid          string `json:"uid"`
	RegisterTime int64  `json:"register_time"`
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

	cli := rpc.GetUserCacheClient()
	if cli == nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	req := &microuser.RegUserInfo{
		Pwd:     hashedPWD,
		Country: int32(data.Param.Country),
		Phone:   data.Param.Phone,
		Type:    int32(constants.LOGIN_TYPE_PHONE),
	}

	resp, err := cli.RegisterUser(context.Background(), req)
	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	if resp.Result != microuser.ResCode_OK {
		response.SetResponseBase(constants.RC_DUP_PHONE)
		return
	}

	response.Data = registerUpVcodeResData{
		utils.Int642Str(resp.Uid),
		resp.RegTime,
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
