package common

import (
	"utils/config"
	"utils/lvthttp"
	"utils/logger"
	"utils"
)

type (
	wxRes struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Openid      string `json:"openid"`
		Scope       string `json:"scope"`
		Unionid     string `json:"unionid"`
	}
)



func AuthWX(code string)(bool,*wxRes){

	wx := config.GetConfig().WXAuth

	param := make(map[string]string,0)
	param["appid"] = wx.Appid
	param["secret"] = wx.Secret
	param["code"] = code
	param["grant_type"] = "authorization_code"

	resBody,err := lvthttp.Get(wx.Url,param)
	if err != nil {
		logger.Error("http req error",err.Error())
		return false,nil
	}

	res := new(wxRes)
	if err = utils.FromJson(resBody,res);err != nil {
		logger.Error("json parse error",err.Error(),"res body",resBody)
		return false,nil
	}
	return true,res
}