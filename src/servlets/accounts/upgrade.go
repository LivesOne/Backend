package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"servlets/token"
)

type upgradeResData struct {
	Level int `json:"level"`
}
// checkVCodeHandler
type upgradeHandler struct {
}

func (handler *upgradeHandler) Method() string {
	return http.MethodPost
}

func (handler *upgradeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)

	if httpHeader.Timestamp < 1 {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}


	if ok,level := common.UserUpgrade(uidString);ok {
		response.Data = upgradeResData{
			Level: level,
		}
	} else {
		response.SetResponseBase(constants.RC_UPGRAD_FAILED)
	}





}
