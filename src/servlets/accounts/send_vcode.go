package accounts

import (
	"encoding/json"
	"github.com/donnie4w/go-logger/logger"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils/config"
	log "utils/logger"
)

const (
	MESSAGE         = 1
	CALL            = 2
	EMAIL           = 3
	MESSAGE_VID     = 2
	VOICE_CODE_SMS  = 0
	VOICE_CODE_CALL = 1
	MAIL_TPL_MAXTHON = 1
	MAIL_TOL_LVT= 2
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
	Expire int `json:"expire"`
}

type sendVCodeRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *sendVCodeParam  `json:"param"`
}

type httpVImgReqParam struct {
	Id   string `json:"id"`
	Code string `json:"code"`
}

type httpReqMessageParam struct {
	AreaCode  int    `json:"area_code"`
	Lan       string `json:"lan"`
	PhoneNo   string `json:"phone_no"`
	Vid       int    `json:"vid"`
	Expire    int    `json:"expire"`
	VoiceCode int    `json:"voice_code"`
}

type httpReqMailParam struct {
	Mail   string `json:"mail"`
	Tpl    int    `json:"tpl"`
	Ln     string `json:"ln"`
	Expire int    `json:"expire"`
}

type httpResSms struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type httpMailVCodeResParam struct {
	Ret int `json:"ret,omitempty"`
	Msg string `json:"msg,omitempty"`
	Data *httpReqVCode `json:"data,omitempty"`
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
	//TODO validate img vcode
	url := config.GetConfig().ImgSvrAddr + "/v/v1/validate"
	typeData := httpVImgReqParam{
		Id:   requestData.Param.IMG_id,
		Code: requestData.Param.IMG_vcode,
	}
	reqParam, _ := json.Marshal(typeData)
	svrResStr, err := common.Post(url, string(reqParam))
	if err != nil {
		log.Error("url ---> ", url, " http send error ", err.Error())
		response.Base = &common.BaseResp{
			RC:  constants.RC_SYSTEM_ERR.Rc,
			Msg: constants.RC_SYSTEM_ERR.Msg,
		}
	} else {
		svrRes := httpVCodeResParam{}
		err := json.Unmarshal([]byte(svrResStr), &svrRes)
		if err != nil {
			log.Info("ParseHttpBodyParams, parse body param error: ", err)
			response.Base = &common.BaseResp{
				RC:  constants.RC_SYSTEM_ERR.Rc,
				Msg: constants.RC_SYSTEM_ERR.Msg,
			}
		}
		if !isNotNull(requestData.Param.Ln) {
			requestData.Param.Ln = "en-us"
		}
		if svrRes.Ret == 0 {
			log.Error("Type",requestData.Param.Type)
			switch requestData.Param.Type {
				case MESSAGE:
					sendMessage(requestData.Param, response)
				case CALL:
					sendCall(requestData.Param, response)
				case EMAIL:
					sendEmail(requestData.Param, response)
			}
		} else {
			response.Base = &common.BaseResp{
				RC:  constants.RC_INVALID_VCODE.Rc,
				Msg: constants.RC_INVALID_VCODE.Msg,
			}
		}
	}
}

func sendMessage(param *sendVCodeParam, res *common.ResponseData) {
	messageServerReq(param, res, VOICE_CODE_SMS)
}

func sendCall(param *sendVCodeParam, res *common.ResponseData) {
	messageServerReq(param, res, VOICE_CODE_CALL)
}

func sendEmail(param *sendVCodeParam, res *common.ResponseData) {
	//TODO send

	if isNotNull(param.EMail) {
		req := httpReqMailParam{
			Mail:   param.EMail,
			Tpl:    MAIL_TOL_LVT,
			Ln:     param.Ln,
			Expire: param.Expire,
		}
		url := config.GetConfig().MailSvrAddr + "/mail/v1/getCode"
		reqStr, _ := json.Marshal(req)
		jsonRes, err := common.Post(url, string(reqStr))
		if err != nil {
			logger.Error("post error ---> ", err.Error())
			res.Base = &common.BaseResp{
				RC:  constants.RC_SYSTEM_ERR.Rc,
				Msg: constants.RC_SYSTEM_ERR.Msg,
			}
		} else {
			svrRes := httpMailVCodeResParam{}
			err := json.Unmarshal([]byte(jsonRes), &svrRes)
			if err != nil {
				log.Info("ParseHttpBodyParams, parse body param error: ", err)
				res.Base = &common.BaseResp{
					RC:  constants.RC_SYSTEM_ERR.Rc,
					Msg: constants.RC_SYSTEM_ERR.Msg,
				}
			}
			if svrRes.Ret == 0 {
				res.Data = &sendVCodeRes{
					Vcode_id:svrRes.Data.Id,
					Expire:svrRes.Data.Expire,
				}
			}

		}
	} else {
		res.Base = &common.BaseResp{
			RC:  constants.RC_PARAM_ERR.Rc,
			Msg: constants.RC_PARAM_ERR.Msg,
		}
	}

}

func messageServerReq(param *sendVCodeParam, res *common.ResponseData, voiceCode int) {
	if isNotNull(param.Phone) && param.Country > 0 {
		//{"area_code":86,"lan":"cn","phone_no":"13901008888","vid":2,"expired":3600,"voice_code":0}

		req := httpReqMessageParam{
			AreaCode:  param.Country,
			Lan:       convSmsLn(param.Ln),
			PhoneNo:   param.Phone,
			Vid:       MESSAGE_VID,
			Expire:    param.Expire,
			VoiceCode: voiceCode,
		}
		url := config.GetConfig().SmsSvrAddr + "/get"
		reqStr, _ := json.Marshal(req)
		jsonRes, err := common.Post(url, string(reqStr))
		if err != nil {
			logger.Error("post error ---> ", err.Error())
			res.Base = &common.BaseResp{
				RC:  constants.RC_SYSTEM_ERR.Rc,
				Msg: constants.RC_SYSTEM_ERR.Msg,
			}
		} else {
			httpRes := httpResSms{}
			json.Unmarshal([]byte(jsonRes), &httpRes)
			if httpRes.Code != 1 {
				res.Base = &common.BaseResp{
					RC:  constants.RC_PARAM_ERR.Rc,
					Msg: constants.RC_PARAM_ERR.Msg,
				}
			}
		}
	} else {
		res.Base = &common.BaseResp{
			RC:  constants.RC_PARAM_ERR.Rc,
			Msg: constants.RC_PARAM_ERR.Msg,
		}
	}
}
func isNotNull(s string) bool {
	return len(s) > 0
}
func convSmsLn(ln string) string {
	if ln == "zh-cn" {
		return "cn"
	} else {
		return "en"
	}
}
