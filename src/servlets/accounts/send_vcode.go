package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
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
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
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
		Data: 0, // data expire Int 失效时间，单位秒
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := sendVCodeRequest{} // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &requestData)
	//validate add exists
	if !validateAction(requestData.Param) {
		response.Base = &common.BaseResp{
			RC:  constants.RC_PARAM_ERR.Rc,
			Msg: "action add params exists",
		}
	} else {
		//validate img vcode

		succFlag, code := vcode.ValidateImgVCode(requestData.Param.IMG_id, requestData.Param.IMG_vcode)
		if succFlag {
			switch requestData.Param.Type {
			case MESSAGE:
				vcode.SendSmsVCode(requestData.Param.Phone, requestData.Param.Country, requestData.Param.Ln, requestData.Param.Expire)
			case CALL:
				vcode.SendCallVCode(requestData.Param.Phone, requestData.Param.Country, requestData.Param.Ln, requestData.Param.Expire)
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
			switch code {
			case vcode.CODE_EXPIRED_ERR:
				response.SetResponseBase(constants.RC_VCODE_EXPIRE)
			case vcode.VALIDATE_CODE_FAILD:
				response.SetResponseBase(constants.RC_INVALID_VCODE)
			default:
				response.Base = &common.BaseResp{
					RC:  constants.RC_PARAM_ERR.Rc,
					Msg: constants.RC_PARAM_ERR.Msg,
				}
			}
		}
	}

}

func validateAction(param *sendVCodeParam) bool {
	if param.Action == "add" {
		switch param.Type {
		case MESSAGE, CALL:
			if common.ExistsPhone(param.Country, param.Phone) {
				return false
			}
		case EMAIL:
			if common.ExistsEmail(param.EMail) {
				return false
			}
		default:
			return false
		}
	}
	return true
}
