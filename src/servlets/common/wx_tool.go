package common

import (
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"servlets/constants"
	"servlets/rpc"
	"strings"
	"utils"
	"utils/config"
	"utils/logger"
	"utils/lvthttp"
)

type (
	wxRes struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Openid      string `json:"openid"`
		Scope       string `json:"scope"`
		Unionid     string `json:"unionid"`
	}
	//{"errcode":41002,"errmsg":"appid missing, hints: [ req_id: a_ZoNA0774th54 ]"}
	wxErrorRes struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
)

func decodeWXres(resBody string) (bool, *wxRes) {
	//校验是否是错误返回的格式
	errRes := new(wxErrorRes)
	if err := utils.FromJson(resBody, errRes); err != nil {
		logger.Error("json parse error", err.Error(), "res body", resBody)
		return false, nil
	}

	if errRes.Errcode > 0 {
		logger.Error("wx req error", errRes.Errmsg)
		return false, nil
	}

	res := new(wxRes)
	if err := utils.FromJson(resBody, res); err != nil {
		logger.Error("json parse error", err.Error(), "res body", resBody)
		return false, nil
	}
	return true, res
}

func buildUrlAndParamsByAuthApp(app, wxCode string) (string, map[string]string) {
	wx := config.GetConfig().WXAuth
	app = strings.ToLower(app)
	var auth config.WXAuthData
	switch app {
	case "mobile":
		auth = wx.Data.Mobile
	default:
		auth = wx.Data.Web
	}
	param := make(map[string]string, 0)
	param["appid"] = auth.Appid
	param["secret"] = auth.Secret
	param["code"] = wxCode
	param["grant_type"] = "authorization_code"
	logger.Info("wx auth url[", wx.Url, "] param[", utils.ToJSON(param), "]")
	return wx.Url, param
}

func AuthWX(app, wxCode string) (bool, *wxRes) {

	url, param := buildUrlAndParamsByAuthApp(app, wxCode)

	resBody, err := lvthttp.Get(url, param)
	if err != nil {
		logger.Error("http req error", err.Error())
		return false, nil
	}
	logger.Info("wx http res ", resBody)

	return decodeWXres(resBody)
}

func SecondAuthWX(uid int64, app, authCode string) (bool, constants.Error) {
	// 微信绑定验证，未绑定返回验提取失败
	wx, _ := rpc.GetUserField(uid, microuser.UserField_WX)
	if len(wx) == 0 {
		logger.Error("user is not bind wx")
		return false, constants.RC_WX_SEC_AUTH_FAILED
	}
	wxIds := strings.Split(wx, ",")
	if len(wxIds) != 2 {
		logger.Error("user is not bind wx")
		return false, constants.RC_WX_SEC_AUTH_FAILED
	}
	openId, unionId := wxIds[0], wxIds[1]
	if len(openId) == 0 || len(unionId) == 0 {
		logger.Error("user is not bind wx")
		return false, constants.RC_WX_SEC_AUTH_FAILED
	}
	//微信认证并比对id
	if ok, res := AuthWX(app, authCode); ok {
		if res.Unionid != unionId {
			logger.Error("user check sec wx failed")
			logger.Error("db openId,unionId [", openId, unionId, "] wx result openId,unionId [", res.Openid, res.Unionid, "]")
			//二次验证不通过扣10分
			//deductionCreditScore := 10
			//log.Error("deduction credit score :",deductionCreditScore)
			//common.DeductionCreditScore(uid,deductionCreditScore)
			return false, constants.RC_WX_SEC_AUTH_FAILED
		}
		return true, constants.RC_OK
	}
	logger.Error("wx auth failed")
	return false, constants.RC_INVALID_WX_CODE
}
