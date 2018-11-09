package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/vcode"
	"utils/logger"
)

type checkVCodeParam struct {
	Type    int    `json:"type"`
	Action  string `json:"action"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	VCode   string `json:"vcode"`
	VCodeId string `json:"vcode_id"`
	Keep    int    `json:"keep"`
}

type checkVCodeRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *checkVCodeParam `json:"param"`
}

// checkVCodeHandler
type checkVCodeHandler struct {
}

func (handler *checkVCodeHandler) Method() string {
	return http.MethodPost
}

func (handler *checkVCodeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	data := checkVCodeRequest{}
	common.ParseHttpBodyParams(request, &data)
	if data.Base == nil || data.Param == nil ||
		(handler.checkRequestParams(header, &data) == false) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	switch data.Param.Type {
	case MESSAGE, CALL:
		f, code := vcode.ValidateSmsAndCallVCode(data.Param.Phone, data.Param.Country, data.Param.VCode, 3600, vcode.FLAG_DEF)
		if !f {
			response.SetResponseBase(vcode.ConvSmsErr(code))
		}
	case EMAIL:
		f, ec := vcode.ValidateMailVCode(data.Param.VCodeId, data.Param.VCode, data.Param.EMail)
		if !f {
			response.SetResponseBase(vcode.ConvImgErr(ec))
		}
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
	}

}

func (handler *checkVCodeHandler) checkRequestParams(header *common.HeaderParams, data *checkVCodeRequest) bool {
	if header == nil || (data == nil) {
		return false
	}

	if header.IsValidTimestamp() == false {
		logger.Info("check verify code: some header param missed")
		return false
	}

	if (data.Base.App == nil) || (data.Base.App.IsValid() == false) {
		logger.Info("check verify code: app info invalid")
		return false
	}

	return true
}
