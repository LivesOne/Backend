package vcode

import (
	"encoding/json"
	"utils"
	"utils/config"
	"utils/logger"
)

const (
	SUCCESS = 0

	NOT_FOUND_ERR = 404

	SERVER_ERR = 500

	NO_PARAMS_ERR = 1

	PARAMS_ERR = 2

	JSON_PARSE_ERR = 3

	CODE_EXPIRED_ERR = 4

	VALIDATE_CODE_FAILD = 5

	EMAIL_VALIDATE_FAILD = 6

	HTTP_ERR = 7

	MESSAGE_VID      = 2
	VOICE_CODE_SMS   = 0
	VOICE_CODE_CALL  = 1
	MAIL_TPL_MAXTHON = 1
	MAIL_TOL_LVT     = 2

	FLAG_KEEP = 1

	FLAG_DEF = 0
)

type httpResParam struct {
	Ret int    `json:"ret,omitempty"`
	Msg string `json:"msg,omitempty"`
}

type httpImgReqParam struct {
	Len    int `json:"len,omitempty"`
	W      int `json:"w,omitempty"`
	H      int `json:"h,omitempty"`
	Expire int `json:"expire,omitempty"`
}

type httpReqVCode struct {
	Expire int    `json:"expire,omitempty"`
	Size   int    `json:"size,omitempty"`
	Id     string `json:"id,omitempty"`
}

type httpReqVCodeData struct {
	ImgBase string        `json:"imgBase,omitempty"`
	VCode   *httpReqVCode `json:"vCode,omitempty"`
}

type httpImgVCodeResParam struct {
	httpResParam
	Data *httpReqVCodeData `json:"data,omitempty"`
}

type httpVReqParam struct {
	Id    string `json:"id,omitempty"`
	Code  string `json:"code,omitempty"`
	Email string `json:"email,omitempty"`
}

type httpReqMessageParam struct {
	AreaCode  int    `json:"area_code"`
	Lan       string `json:"lan"`
	PhoneNo   string `json:"phone_no"`
	Vid       int    `json:"vid"`
	Expire    int    `json:"expired"`
	VoiceCode int    `json:"voice_code"`
}

type httpReqValidateMessageParam struct {
	AreaCode       int    `json:"area_code"`
	ValidationCode string `json:"validation_code"`
	PhoneNo        string `json:"phone_no"`
	Vid            int    `json:"vid"`
	Flag           string `json:"flag"`
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
	Ret  int           `json:"ret,omitempty"`
	Msg  string        `json:"msg,omitempty"`
	Data *httpReqVCode `json:"data,omitempty"`
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

func GetImgVCode(w, h, len, expire int) *httpImgVCodeResParam {
	typeData := httpImgReqParam{
		W:      w,
		H:      h,
		Len:    len,
		Expire: expire,
	}
	reqParam, _ := json.Marshal(typeData)
	url := config.GetConfig().ImgSvrAddr + "/img/v1/getCode"
	svrResStr, err := utils.Post(url, string(reqParam))
	if err != nil {
		logger.Error("url ---> ", url, " http send error ", err.Error())
	} else {
		svrRes := httpImgVCodeResParam{}
		err := json.Unmarshal([]byte(svrResStr), &svrRes)
		if err != nil {
			logger.Info("ParseHttpBodyParams, parse body param error: ", err)
		}
		return &svrRes
	}
	return nil
}

func messageServerReq(phone string, country int, ln string, expire int, voiceCode int) (bool, error) {
	if isNotNull(phone) && country > 0 {
		//{"area_code":86,"lan":"cn","phone_no":"13901008888","vid":2,"expired":3600,"voice_code":0}

		req := httpReqMessageParam{
			AreaCode:  country,
			Lan:       convSmsLn(ln),
			PhoneNo:   phone,
			Vid:       MESSAGE_VID,
			Expire:    expire,
			VoiceCode: voiceCode,
		}
		url := config.GetConfig().SmsSvrAddr + "/get"
		reqStr, _ := json.Marshal(req)
		jsonRes, err := utils.Post(url, string(reqStr))
		if err != nil {
			logger.Error("post error ---> ", err.Error())
			return false, err
		} else {
			httpRes := httpResSms{}
			json.Unmarshal([]byte(jsonRes), &httpRes)
			return httpRes.Code == 1, nil
		}
	} else {

		return false, nil
	}
}

func SendSmsVCode(phone string, country int, ln string, expire int) (bool, error) {
	return messageServerReq(phone, country, ln, expire, VOICE_CODE_SMS)
}

func SendCallVCode(phone string, country int, ln string, expire int) (bool, error) {
	return messageServerReq(phone, country, ln, expire, VOICE_CODE_CALL)
}



func SendMailVCode(email string, ln string, expire int) *httpReqVCode {
	if isNotNull(email) {
		req := httpReqMailParam{
			Mail:   email,
			Tpl:    MAIL_TOL_LVT,
			Ln:     ln,
			Expire: expire,
		}
		url := config.GetConfig().MailSvrAddr + "/mail/v1/getCode"
		reqStr, _ := json.Marshal(req)
		jsonRes, err := utils.Post(url, string(reqStr))
		if err != nil {
			logger.Error("post error ---> ", err.Error())
		} else {
			svrRes := httpMailVCodeResParam{}
			err := json.Unmarshal([]byte(jsonRes), &svrRes)
			if err != nil {
				logger.Info("ParseHttpBodyParams, parse body param error: ", err)
			}
			return svrRes.Data

		}
	} else {
		logger.Error("email can not be null")
	}
	return nil
}

func ValidateImgVCode(id string, vcode string) (bool, int) {
	url := config.GetConfig().ImgSvrAddr + "/v/v1/validate"
	typeData := httpVReqParam{
		Id:   id,
		Code: vcode,
	}
	reqParam, _ := json.Marshal(typeData)
	svrResStr, err := utils.Post(url, string(reqParam))
	if err != nil {
		logger.Error("url ---> ", url, " http send error ", err.Error())
		return false, HTTP_ERR
	}
	svrRes := httpResParam{}
	err1 := json.Unmarshal([]byte(svrResStr), &svrRes)
	if err1 != nil {
		logger.Info("ParseHttpBodyParams, parse body param error: ", err)
		return false, JSON_PARSE_ERR
	}
	return svrRes.Ret == SUCCESS, svrRes.Ret
}

func ValidateSmsAndCallVCode(phone string, country int, code string, expire int, flag int) (bool, error) {
	if isNotNull(phone) && country > 0 {
		//{"area_code":86,"lan":"cn","phone_no":"13901008888","vid":2,"expired":3600,"voice_code":0}

		req := httpReqValidateMessageParam{
			AreaCode:       country,
			ValidationCode: code,
			PhoneNo:        phone,
			Vid:            MESSAGE_VID,
			Flag:           utils.Int2Str(flag),
		}
		url := config.GetConfig().SmsSvrAddr + "/validate"
		reqStr, _ := json.Marshal(req)
		jsonRes, err := utils.Post(url, string(reqStr))
		if err != nil {
			logger.Error("post error ---> ", err.Error())
			return false, err
		} else {
			httpRes := httpResSms{}
			json.Unmarshal([]byte(jsonRes), &httpRes)
			return httpRes.Code == 1, nil
		}
	} else {
		return false, nil
	}
}

func ValidateMailVCode(id string, vcode string, email string) (bool, int) {
	if len(id) > 0 && len(vcode) > 0 && len(email) > 0 {
		url := config.GetConfig().ImgSvrAddr + "/v/v1/validate"
		typeData := httpVReqParam{
			Id:    id,
			Code:  vcode,
			Email: email,
		}
		reqParam, _ := json.Marshal(typeData)
		svrResStr, err := utils.Post(url, string(reqParam))
		if err != nil {
			logger.Error("url ---> ", url, " http send error ", err.Error())
			return false, HTTP_ERR
		}
		svrRes := httpResParam{}
		err1 := json.Unmarshal([]byte(svrResStr), &svrRes)
		if err1 != nil {
			logger.Info("ParseHttpBodyParams, parse body param error: ", err)
			return false, JSON_PARSE_ERR
		}
		return svrRes.Ret == SUCCESS, svrRes.Ret
	} else {
		logger.Error("vcode_id||vcode||email can not be empty")
		return false, PARAMS_ERR
	}
}
