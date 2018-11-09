package vcode

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"gitlab.maxthon.net/cloud/base-sms-gateway/src/proto"
	"gitlab.maxthon.net/cloud/base-vcode/src/proto"
	"math/rand"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
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

	SMS_UP_URL_PATH       = "/sms/v1/validate"
	IMG_URL_PATH          = "/img/v1/getCode"
	SMS_URL_PATH          = "/get"
	MAIL_URL_PATH         = "/mail/v1/getCode"
	VALIDATE_URL_PATH     = "/v/v1/validate"
	SMS_VALIDATE_URL_PATH = "/validate"
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
	upSmsReq struct {
		Country int    `json:"country"`
		Phone   string `json:"mobile"`
		Code    string `json:"code"`
	}
	ImgMailRes struct {
		Id     string
		Code   int
		Size   int
		Data   string
		Expire int
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

func messageServerReq(phone string, country int, ln string, expire int, voiceCode int) (bool, error) {
	if isNotNull(phone) && country > 0 {
		cli := rpc.GetSmsClient()
		if cli != nil {
			req := &smspb.VoiceMsgRequest{
				Country:   int32(country),
				Lan:       convSmsLn(ln),
				Phone:     phone,
				Vid:       MESSAGE_VID,
				Expired:   int32(expire),
				VoiceCode: int32(voiceCode),
			}
			resp, err := cli.SmsSendVoiceMsg(context.Background(), req)
			if err != nil {
				logger.Error("grpc SmsSendVoiceMsg request error: ", err)
				return false, err
			} else {
				return resp.Code == 1, nil
			}
		}
	}
	return false, nil
}

func SendSmsVCode(phone string, country int, ln string, expire int) (bool, error) {
	return messageServerReq(phone, country, ln, expire, VOICE_CODE_SMS)
}

func SendCallVCode(phone string, country int, ln string, expire int) (bool, error) {
	return messageServerReq(phone, country, ln, expire, VOICE_CODE_CALL)
}

func GetImgVCode(w, h, len, expire int) (*ImgMailRes, error) {
	cli := rpc.GetVcodeClient()
	if cli != nil {
		reqData := &vcodeproto.ImgVcodeReq{
			W:      int32(w),
			H:      int32(h),
			Len:    int32(len),
			Expire: int32(expire),
		}
		resp, err := cli.SendImgVcode(context.Background(), reqData)
		if err != nil || resp == nil {
			logger.Error("grpc SendImgVcode request error: ", err)
			return nil, err
		} else {
			return &ImgMailRes{
				Id:     resp.Data.Vcode.Id,
				Code:   int(resp.Code),
				Size:   int(resp.Data.Vcode.Size),
				Data:   resp.Data.ImgBase,
				Expire: int(resp.Data.Vcode.Expire),
			}, nil
		}
	}
	return nil, errors.New("can not conn to rpc service")
}

func SendMailVCode(email string, ln string, expire int) (*ImgMailRes, error) {
	cli := rpc.GetVcodeClient()
	if cli != nil {
		reqData := &vcodeproto.EmailVcodeReq{
			Mail:   email,
			Ln:     ln,
			Tpl:    MAIL_TOL_LVT,
			Expire: int32(expire),
		}
		resp, err := cli.SendEmailVcode(context.Background(), reqData)
		if err != nil || resp == nil {
			logger.Error("grpc SendEmailVcode request error: ", err)
			return nil, err
		} else {
			return &ImgMailRes{
				Id:     resp.Data.Id,
				Code:   int(resp.Code),
				Size:   int(resp.Data.Size),
				Expire: int(resp.Data.Expire),
			}, nil
		}
	}
	return nil, errors.New("can not conn to rpc service")
}

func validateImgVCode(id string, vcode string) (bool, int) {
	cli := rpc.GetVcodeClient()
	if cli != nil {
		reqData := &vcodeproto.ValidateRequest{
			Id:   id,
			Code: vcode,
			Vm:   0,
		}
		resp, err := cli.ValidateVcode(context.Background(), reqData)
		if err != nil {
			logger.Error("grpc ValidateVcode request error: ", err)
			return false, HTTP_ERR
		} else {
			return resp.Code == SUCCESS, int(resp.Code)
		}
	}
	return false, HTTP_ERR
}

func ValidateImgVCode(id, vcode string) (bool, int) {
	if len(id) > 0 && len(vcode) > 0 {
		return validateImgVCode(id, vcode)
	} else if len(id) == 0 && len(vcode) > 0 {
		return ValidateWYYD(vcode)
	}
	return false, PARAMS_ERR
}

func ValidateSmsAndCallVCode(phone string, country int, code string, expire int, flag int) (bool, int) {
	if isNotNull(phone) && country > 0 {
		cli := rpc.GetSmsClient()
		if cli != nil {
			req := &smspb.ValidateRequest{
				Country:        int32(country),
				Phone:          phone,
				Flag:           utils.Int2Str(flag),
				ValidationCode: code}
			resp, err := cli.SmsValidate(context.Background(), req)
			if err != nil {
				logger.Error("grpc SmsSendMsg request error: ", err)
				return false, SMS_PROTOCOL_ERR
			} else {
				return resp.Code == SMS_SUCC, int(resp.Code)
			}
		}
	}
	return false, SMS_PROTOCOL_ERR
}

func ValidateMailVCode(id string, vcode string, email string) (bool, int) {
	if len(id) > 0 && len(vcode) > 0 && len(email) > 0 {
		cli := rpc.GetVcodeClient()
		if cli != nil {
			reqData := &vcodeproto.ValidateRequest{
				Id:    id,
				Code:  vcode,
				Email: email,
				Vm:    0,
			}
			resp, err := cli.ValidateVcode(context.Background(), reqData)
			if err != nil {
				logger.Error("grpc ValidateVcode request error: ", err)
				return false, HTTP_ERR
			} else {
				return resp.Code == SUCCESS, int(resp.Code)
			}
		}
		return false, HTTP_ERR
	}
	logger.Error("vcode_id||vcode||email can not be empty")
	logger.Error("id --> ", id, " code --> ", vcode, " email --> ", email)
	return false, PARAMS_ERR
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
		captcha := config.GetConfig().Captcha
		ts := utils.GetTimestamp13()
		rand.Seed(ts)
		param := make(map[string]string, 0)
		param["captchaId"] = captcha.Id
		param["validate"] = validate
		param["user"] = ""
		param["secretId"] = captcha.SecretId
		param["version"] = "v2"
		param["timestamp"] = utils.Int642Str(ts)
		param["nonce"] = utils.Int2Str(rand.Intn(200))
		param["signature"] = genSignature(captcha.SecretKey, param)
		logger.Debug("validate req params --->", utils.ToJSON(param))
		resBodyStr, err := lvthttp.FormPost(captcha.Url, param)
		if err != nil {
			logger.Error("http req eror ---> ", err.Error())
			return false, HTTP_ERR
		}
		logger.Debug("validate response --->", resBodyStr)
		res := wyydRes{}
		err1 := utils.FromJson(resBodyStr, &res)
		if err1 != nil {
			logger.Info("ParseHttpBodyParams, parse body param error: ", err)
			return false, JSON_PARSE_ERR
		}
		ret := 0

		if !res.Result {
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

func ValidateSmsUpVCode(country int, phone, code string) (bool, constants.Error) {
	url := config.GetConfig().SmsUpValidateSvrAddr + SMS_UP_URL_PATH
	sms := upSmsReq{
		Country: country,
		Phone:   phone,
		Code:    code,
	}
	resp, err := lvthttp.JsonPost(url, sms)
	if err != nil {
		return false, constants.RC_SYSTEM_ERR
	}
	res := new(common.ResponseData)
	if err := utils.FromJson(resp, res); err != nil {
		return false, constants.RC_SYSTEM_ERR
	}
	base := res.Base
	return base.RC == 0, constants.Error{base.RC, base.Msg}
}
