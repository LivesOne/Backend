package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils/logger"
	"utils/vcode"
)

const (
	MESSAGE = 1
	CALL    = 2
	EMAIL   = 3
)

type sendVCodeParam struct {
	IMG_id      string `json:"img_id"`
	IMG_vcode   string `json:"img_vcode"`
	Type        int    `json:"type"`
	Action      string `json:"action"`
	Country     int    `json:"country"`
	Phone       string `json:"phone"`
	Check_phone int    `json:"check_phone"`
	EMail       string `json:"email"`
	Ln          string `json:"ln"`
	Expire      int    `json:"expire"`
}

type sendVCodeRes struct {
	Vcode_id string `json:"vcode_id"`
	Expire   int    `json:"expire"`
}

type sendVCodeRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *sendVCodeParam  `json:"param"`
}

// sendVCodeHandler
type sendVCodeHandler struct {
}

func (handler *sendVCodeHandler) Method() string {
	return http.MethodPost
}

func (handler *sendVCodeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	requestData := sendVCodeRequest{} // request body
	common.ParseHttpBodyParams(request, &requestData)

	if requestData.Base == nil || requestData.Param == nil ||
		(handler.checkRequestParams(header, &requestData) == false) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	//validate add exists
	validateFlag, rcErr := validateAction(requestData.Param)
	if !validateFlag {
		response.SetResponseBase(rcErr)
	} else {
		//validate img vcode

		succFlag, code := vcode.ValidateImgVCode(requestData.Param.IMG_id, requestData.Param.IMG_vcode)
		if succFlag {
			switch requestData.Param.Type {
			case MESSAGE:
				f, _ := vcode.SendSmsVCode(requestData.Param.Phone, requestData.Param.Country, requestData.Param.Ln, requestData.Param.Expire)
				if f {
					response.Data = &sendVCodeRes{
						Vcode_id: "maxthonVCodeId",
						Expire:   requestData.Param.Expire,
					}
				} else {
					response.Base = &common.BaseResp{
						RC:  constants.RC_PARAM_ERR.Rc,
						Msg: constants.RC_PARAM_ERR.Msg,
					}
				}
			case CALL:
				f, _ := vcode.SendCallVCode(requestData.Param.Phone, requestData.Param.Country, requestData.Param.Ln, requestData.Param.Expire)
				if f {
					response.Data = &sendVCodeRes{
						Vcode_id: "maxthonVCodeId",
						Expire:   requestData.Param.Expire,
					}
				} else {
					response.Base = &common.BaseResp{
						RC:  constants.RC_PARAM_ERR.Rc,
						Msg: constants.RC_PARAM_ERR.Msg,
					}
				}
			case EMAIL:
				svrRes := vcode.SendMailVCode(requestData.Param.EMail, requestData.Param.Ln, requestData.Param.Expire)
				if svrRes != nil {
					response.Data = &sendVCodeRes{
						Vcode_id: svrRes.Id,
						Expire:   svrRes.Expire,
					}
				} else {
					response.Base = &common.BaseResp{
						RC:  constants.RC_PARAM_ERR.Rc,
						Msg: constants.RC_PARAM_ERR.Msg,
					}
				}
			}
		} else {
			response.SetResponseBase(vcode.ConvImgErr(code))
		}
	}

}

func validateAction(param *sendVCodeParam) (bool, constants.Error) {
	if param.Action == "add" {
		switch param.Type {
		case MESSAGE, CALL:
			if common.ExistsPhone(param.Country, param.Phone) {
				return false, constants.RC_DUP_PHONE
			}
		case EMAIL:
			if common.ExistsEmail(param.EMail) {
				return false, constants.RC_DUP_EMAIL
			}
		default:
			return false, constants.RC_PARAM_ERR
		}
	} else if param.Action == "reset" {
		switch param.Type {
		case MESSAGE, CALL:
			if common.CheckResetPhone(param.Country, param.Phone) {
				return false, constants.RC_INVALID_ACCOUNT
			}
		case EMAIL:
			if common.CheckResetEmail(param.EMail) {
				return false, constants.RC_INVALID_ACCOUNT
			}
		default:
			return false, constants.RC_PARAM_ERR
		}
	}
	return true, constants.RC_OK
}

func (handler *sendVCodeHandler) checkRequestParams(header *common.HeaderParams, data *sendVCodeRequest) bool {
	if header == nil || (data == nil) {
		return false
	}

	if header.IsValidTimestamp() == false {
		logger.Info("send verify code: some header param missed")
		return false
	}

	if (data.Base.App == nil) || (data.Base.App.IsValid() == false) {
		logger.Info("send verify code: app info invalid")
		return false
	}

	return true
}
