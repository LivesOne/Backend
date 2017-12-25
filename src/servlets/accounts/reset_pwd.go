package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/vcode"
	"utils/config"
)

type resetPwdParam struct {
	Type    int    `json:"type"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	VCodeId string `json:"vcode_id"`
	VCode   string `json:"vcode"`
	PWD     string `json:"pwd"`
	Spkv    int    `json:"spkv"`
}

type resetPwdRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param resetPwdParam   `json:"param"`
}

// resetPwdHandler
type resetPwdHandler struct {
	header      *common.HeaderParams // request header param
	requestData *resetPwdRequest     // request body
}

func (handler *resetPwdHandler) Method() string {
	return http.MethodPost
}

func (handler *resetPwdHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)
	requestData := new(resetPwdRequest)
	common.ParseHttpBodyParams(request, &requestData)

	if httpHeader.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, _, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
	}
	uid := utils.Str2Int64(uidString)
	account, err := common.GetAccountByUID(uidString)
	if err != nil {
		response.SetResponseBase(constants.RC_INVALID_ACCOUNT)
		return
	}

	// 检查验证码
	checkType := requestData.Param.Type
	if checkType == 1 {
		ok, err := vcode.ValidateMailVCode(
			requestData.Param.VCodeId, requestData.Param.VCode, account.Email)
		if ok == false {
			response.SetResponseBase(ValidateMailVCodeErr2RcErr(err))
			return
		}

	} else if checkType == 2 {
		ok, err := vcode.ValidateSmsAndCallVCode(
			account.Phone, account.Country, requestData.Param.VCode,0, 0)
		if err != nil || ok == false {
			response.SetResponseBase(constants.RC_INVALID_VCODE)
			return
		}

	} else {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 解析出密码哈希
	pwd, err := utils.RsaDecrypt(requestData.Param.PWD, config.GetPrivateKey())
	if err != nil {
		response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		return
	}

	// save to db
	if err := common.SetLoginPassword(uid, pwd); err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
