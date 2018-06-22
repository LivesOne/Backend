package common

import (
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

func AuthWX(code string) (bool, *wxRes) {

	wx := config.GetConfig().WXAuth

	param := make(map[string]string, 0)
	param["appid"] = wx.Appid
	param["secret"] = wx.Secret
	param["code"] = code
	param["grant_type"] = "authorization_code"

	resBody, err := lvthttp.Get(wx.Url, param)
	if err != nil {
		logger.Error("http req error", err.Error())
		return false, nil
	}
	logger.Info("wx http res ", resBody)

	//校验是否是错误返回的格式
	errRes := new(wxErrorRes)
	if err = utils.FromJson(resBody, errRes); err != nil {
		logger.Error("json parse error", err.Error(), "res body", resBody)
		return false, nil
	}

	if errRes.Errcode > 0 {
		logger.Error("wx req error", errRes.Errmsg)
		return false, nil
	}

	res := new(wxRes)
	if err = utils.FromJson(resBody, res); err != nil {
		logger.Error("json parse error", err.Error(), "res body", resBody)
		return false, nil
	}
	return true, res
}
