package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
)

type rechargeMemoRespData struct {
	Memo string `json:"memo,omitempty"`
}

// bindEMailHandler
type rechargeMemoHandler struct {
}

func (handler *rechargeMemoHandler) Method() string {
	return http.MethodPost
}

func (handler *rechargeMemoHandler) Handle(
	request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)
	if httpHeader.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uid, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	respData := new(rechargeMemoRespData)
	respData.Memo = common.GenerateMemoFromUID(uid)
	response.Data = respData
	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}
