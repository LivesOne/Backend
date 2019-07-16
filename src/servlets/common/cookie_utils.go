package common

import (
	"strings"
	"utils"
	"utils/config"
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

func GetCookieByTokenAndKey(token,key,uid string)(string,error){

	cc := strings.Replace(getDefautKey(),"$key",uid,-1)
	str := token + "_" + key
	at := utils.NewAesTool([]byte(cc),16)
	b,e := at.Encrypt([]byte(str))
	if e == nil {
		return string(b),nil
	}
	return "",e
}