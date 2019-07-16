package common

import (
	"errors"
	"servlets/constants"
	"utils"
	"utils/config"
	"utils/logger"
)


var (
	baseKey string
	baseIV string
)


func getDefautKey()(string,string){
	if len(baseKey) == 0 {
		cookieCert := config.GetConfig().CookieCert
		if len(cookieCert) ==  constants.AES_totalLen {
			baseKey,baseIV = cookieCert[:constants.AES_ivLen],cookieCert[constants.AES_ivLen:]
		} else {
			logger.Error("can not load cooike cret",cookieCert)
		}
	}
	return baseKey,baseIV
}

func GetCookieByTokenAndKey(token,key string)(string,error){
	if len(token) == 0 || len(key) !=  constants.AES_totalLen {
		return "",errors.New("wrong token or key")
	}
	baseKey,baseIv := getDefautKey()
	str := token + "_" + key
	return utils.AesEncrypt(str,baseKey,baseIv)
}