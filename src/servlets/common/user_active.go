package common

import (
	"utils/logger"
	"utils"
)

const(
	USER_ACTIVE_REDIS_KEY_PROXY = "user:active:"
)

func queryUserActiveDays(uid int64)(bool,int){
	c,days,ts := 0,0,int64(0)
	tx,err := gDbUser.Begin()
	if err != nil {
		logger.Error("begin tx error",err.Error())
		return false,0
	}
	row := tx.QueryRow("select count(1) as c from account_extend where uid = ?",uid)
	if err := row.Scan(&c);err != nil {
		tx.Rollback()
		return false,0
	}

	if c == 0 {
		tx.Exec("insert into account_extend (uid,credit_score) values (?,70)",uid)
		days = MoveMinerDays(uid)
	}else{
		row := tx.QueryRow("select active_days,update_time from account_extend where uid = ?",uid)
		if err := row.Scan(&days,&ts);err != nil {
			tx.Rollback()
			return false,0
		}
		if days == 0 {
			days = MoveMinerDays(uid)
		} else {
			days ++
		}
	}
	_,err = tx.Exec("update account_extend set active_days = ?,update_time = ? where uid = ?",days,utils.GetTimestamp13(),uid)
	if err != nil {
		logger.Error("update user active_days error",err.Error())
		tx.Rollback()
		return false,0
	}
	if err := tx.Commit();err != nil {
		logger.Error("commit tx error",err.Error())
		return false,0
	}
	return true,days
}


func ActiveUser(uid int64){
	key := USER_CACHE_REDIS_KEY_PROXY + utils.Int642Str(uid)
	c,e := ttl(key)
	if e !=nil {
		logger.Error("ttl redis error",e.Error())
		return
	}
	if c > 0 {
		logger.Info("key",key,"exists expire",c)
		return
	}

	if ok,days := queryUserActiveDays(uid);ok{
		SetCacheUserField(uid,USER_CACHE_REDIS_FIELD_NAME_ACTIVE_DAYS,utils.Int2Str(days))
		setAndExpire(key,1,int(utils.GetTomorrowStartTs10()))
	}
}

func QueryUserActiveDaysByCache(uid int64) int64 {
	dayStr, err := GetCacheUserField(uid, USER_CACHE_REDIS_FIELD_NAME_ACTIVE_DAYS)
	if err != nil {
		logger.Info("get uid:", uid, " nick name err,", err)
		return 0
	}
	return utils.Str2Int64(dayStr)
}