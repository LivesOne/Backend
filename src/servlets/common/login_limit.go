package common

import (
	"utils"
	"utils/logger"
	"utils/config"
)

const (
	LOGIN_LIMIT_REDIS_PROXY = "login_limit_"
	PWD_ERR_REDIS_PROXY     = "pwd_err_"
)

func AddWrongPwd(uid int64) (bool,int){
	key := PWD_ERR_REDIS_PROXY + utils.Int642Str(uid)
	inc, err := incr(key)
	if err != nil {
		logger.Error("redis incr error", err.Error())
		return false,0
	}
	if inc == 1 {
		rdsExpire(key, DAY_S)
	}

	c,min := 0,0

	for _,v := range config.GetConfig().LoginPwdErrCntLimit {
		if inc >= v.Number && c < v.Number {
			c = v.Number
			min = v.Min
		}
	}

	if c >0 && min > 0 {
		setUserLimt(uid, min*60)
		return true,min*60
	}
	return false,0
}

func CheckUserInLoginLimit(uid int64) (bool, int) {
	key := LOGIN_LIMIT_REDIS_PROXY + utils.Int642Str(uid)

	expire, err := ttl(key)
	if err != nil {
		logger.Error("redis ttl error", err.Error())
	}
	flag := expire > 0

	return flag, expire
}

func setUserLimt(uid int64, expire int) {
	key := LOGIN_LIMIT_REDIS_PROXY + utils.Int642Str(uid)
	err := setAndExpire(key, 1, expire)
	if err != nil {
		logger.Error("redis set error", err.Error())
	}
}


func ClearUserLimitNum(uid int64){
	key := PWD_ERR_REDIS_PROXY + utils.Int642Str(uid)
	err := rdsDel(key)
	if err != nil {
		logger.Error("redis del error",err.Error())
	}
}