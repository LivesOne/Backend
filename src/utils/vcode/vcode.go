package vcode

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"servlets/constants"
	"sort"
	"utils"
	"utils/config"
	"utils/logger"
	"utils/lvthttp"
)

const (
	SUCCESS = 0

	NOT_FOUND_ERR        = 404
	SERVER_ERR           = 500
	NO_PARAMS_ERR        = 1
	PARAMS_ERR           = 2
	JSON_PARSE_ERR       = 3
	CODE_EXPIRED_ERR     = 4
	VALIDATE_CODE_FAILD  = 5
	EMAIL_VALIDATE_FAILD = 6

	HTTP_ERR = 7

	MESSAGE_VID      = 2
	VOICE_CODE_SMS   = 0
	VOICE_CODE_CALL  = 1
	MAIL_TPL_MAXTHON = 1
	MAIL_TOL_LVT     = 2

	FLAG_KEEP = 1

	FLAG_DEF = 0

	SMS_SUCC                = 1
	SMS_PROTOCOL_ERR        = 200
	SMS_CODE_EXPIRED_ERR    = 103
	SMS_VALIDATE_CODE_FAILD = 102
)

type (
	httpResParam struct {
		Ret int    `json:"ret,omitempty"`
		Msg string `json:"msg,omitempty"`
	}

	httpImgReqParam struct {
		Len    int `json:"len,omitempty"`
		W      int `json:"w,omitempty"`
		H      int `json:"h,omitempty"`
		Expire int `json:"expire,omitempty"`
	}

	httpReqVCode struct {
		Expire int    `json:"expire,omitempty"`
		Size   int    `json:"size,omitempty"`
		Id     string `json:"id,omitempty"`
	}

	httpReqVCodeData struct {
		ImgBase string        `json:"imgBase,omitempty"`
		VCode   *httpReqVCode `json:"vCode,omitempty"`
	}

	httpImgVCodeResParam struct {
		httpResParam
		Data *httpReqVCodeData `json:"data,omitempty"`
	}

	httpVReqParam struct {
		Id    string `json:"id,omitempty"`
		Code  string `json:"code,omitempty"`
		Email string `json:"email,omitempty"`
	}

	httpReqMessageParam struct {
		AreaCode  int    `json:"area_code"`
		Lan       string `json:"lan"`
		PhoneNo   string `json:"phone_no"`
		Vid       int    `json:"vid"`
		Expire    int    `json:"expired"`
		VoiceCode int    `json:"voice_code"`
	}

	httpReqValidateMessageParam struct {
		AreaCode       int    `json:"area_code"`
		ValidationCode string `json:"validation_code"`
		PhoneNo        string `json:"phone_no"`
		Vid            int    `json:"vid"`
		Flag           string `json:"flag"`
	}

	httpReqMailParam struct {
		Mail   string `json:"mail"`
		Tpl    int    `json:"tpl"`
		Ln     string `json:"ln"`
		Expire int    `json:"expire"`
	}

	httpResSms struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	httpMailVCodeResParam struct {
		Ret  int           `json:"ret,omitempty"`
		Msg  string        `json:"msg,omitempty"`
		Data *httpReqVCode `json:"data,omitempty"`
	}
	wyydRes struct {
		Result bool   `json:"result"`
		Error  int    `json:"error"`
		Msg    string `json:"msg"`
	}
)

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

func validateImgVCode(id string, vcode string) (bool, int) {
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

func ValidateImgVCode(id, vcode string) (bool, int) {
	if len(id) > 0 && len(vcode) > 0 {
		return validateImgVCode(id, vcode)
	} else if len(id) == 0 && len(vcode) > 0 {
		return ValidateWYYD(vcode)
	}
	return false,PARAMS_ERR
}

func ValidateSmsAndCallVCode(phone string, country int, code string, expire int, flag int) (bool, int) {
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
			return false, SMS_PROTOCOL_ERR
		} else {
			httpRes := httpResSms{}
			json.Unmarshal([]byte(jsonRes), &httpRes)
			return httpRes.Code == SMS_SUCC, httpRes.Code
		}
	} else {
		return false, SMS_PROTOCOL_ERR
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
		logger.Error("id --> ", id, " code --> ", vcode, " email --> ", email)
		return false, PARAMS_ERR
	}
}

func genSignature(secretKey string, params map[string]string) string {
	var keys []string
	for key, _ := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	buf := bytes.NewBufferString("")
	for _, key := range keys {
		buf.WriteString(key + params[key])
	}
	buf.WriteString(secretKey)
	has := md5.Sum(buf.Bytes())
	return fmt.Sprintf("%x", has)
}

func ValidateWYYD(validate string) (bool, int) {
	if len(validate) > 0 {
		ts := utils.GetTimestamp13()
		rand.Seed(ts)
		param := make(map[string]string, 0)
		param["captchaId"] = config.GetConfig().CAPTCHA_ID
		param["validate"] = validate
		param["user"] = ""
		param["secretId"] = config.GetConfig().CAPTCHA_SECRET_ID
		param["version"] = "v2"
		param["timestamp"] = utils.Int642Str(ts)
		param["nonce"] = utils.Int2Str(rand.Intn(200))
		param["signature"] = genSignature(config.GetConfig().CAPTCHA_SECRET_KEY, param)
		resBodyStr, err := lvthttp.FormPost(config.GetConfig().CAPTCHA_URL, param)
		if err != nil {
			return false, HTTP_ERR
		}
		res := wyydRes{}
		err1 := json.Unmarshal([]byte(resBodyStr), &res)
		if err1 != nil {
			logger.Info("ParseHttpBodyParams, parse body param error: ", err)
			return false, JSON_PARSE_ERR
		}
		ret := 0

		if ! res.Result {
			switch res.Error {
			case 419:
				ret = VALIDATE_CODE_FAILD
			case 415:
				ret = SERVER_ERR
			}
		}

		return res.Result, ret
	}
	return false, PARAMS_ERR
}

func ConvImgErr(code int) constants.Error {
	switch code {
	case CODE_EXPIRED_ERR:
		return constants.RC_VCODE_EXPIRE
	case VALIDATE_CODE_FAILD:
		return constants.RC_INVALID_VCODE
	case EMAIL_VALIDATE_FAILD:
		return constants.RC_EMAIL_NOT_MATCH
	case PARAMS_ERR:
		return constants.RC_PROTOCOL_ERR
	default:
		return constants.RC_SYSTEM_ERR
	}
}

func ConvSmsErr(code int) constants.Error {
	switch code {
	case SMS_CODE_EXPIRED_ERR:
		return constants.RC_VCODE_EXPIRE
	case SMS_VALIDATE_CODE_FAILD:
		return constants.RC_INVALID_VCODE
	case EMAIL_VALIDATE_FAILD:
		return constants.RC_EMAIL_NOT_MATCH
	case SMS_PROTOCOL_ERR:
		return constants.RC_PROTOCOL_ERR
	default:
		return constants.RC_SYSTEM_ERR
	}
}
