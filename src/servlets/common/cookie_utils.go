package common

import (
	"net/url"
	"strings"
	"utils"
	"utils/config"
	"utils/logger"
)


var (
	cookieCert string
)


func getDefautKey()(string){
	if len(cookieCert) == 0 {
		cookieCert = config.GetConfig().CookieCert
	}
	return cookieCert
}

func GetCookieByTokenAndKey(token,key,uid string)(string){

	cc := strings.Replace(getDefautKey(),"$key",uid,-1)
	str := token + "_" + key
	logger.Info("aes ecb src",str,"key",cc)
	at := utils.New256ECBEncrypter(cc)
	return url.QueryEscape(url.QueryEscape(at.Crypt(str)))
}


