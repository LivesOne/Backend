package common

import (
	"github.com/garyburd/redigo/redis"
	"utils/logger"
	"servlets/constants"
	"utils"
	"errors"
	"utils/config"
)

const(
	DAILY_PREPARE_KEY_PROXY = "tl:dp:"
	DAILY_COMMIT_KEY_PROXY = "tl:dc:"
	DAILY_TOTAL_TRANSFER_KEY_PROXY = "tl:dt:"
	USER_TRANS_KEY_PROXY = "tx:uid:"
	TS = 1000
	DAY_S = 24*3600
	DAY_30 = DAY_S*30
	DAY_TS = DAY_S*TS
	LVT_CONV = 100000000
)

var cfg map[int]config.TransferLimit

func getCFG(userLevel int)*config.TransferLimit{
	if cfg == nil {
		cfg = config.GetConfig().TransferLimit
	}
	if lim,ok := cfg[userLevel];ok{
		return &lim
	}
	return nil
}

func rdsDo(commandName string, args ...interface{})(reply interface{}, err error){
	conn := GetRedisConn()
	if conn == nil {
		return 0,errors.New("can not connect redis")
	}
	defer conn.Close()
	return conn.Do(commandName,args...)
}

func ttl(key string)(int,error){
	return redis.Int(rdsDo("TTL",key))
}

func incr(key string)(int ,error){
	return redis.Int(rdsDo("INCR",key))
}

func incrby(key string,value int64)(int ,error){
	return redis.Int(rdsDo("INCRBY",key,value))
}

func rdsGet(key string)(int,error){
	return redis.Int(rdsDo("GET",key))
}


func setAndExpire(key string,value,expire int)error{
	_,err := rdsDo("SET",key,value,"EX",expire)
	return err
}


func checkLimit(key string,limit int,incrFlag bool)(bool,constants.Error){
	t,e :=  ttl(key)
	if e != nil {
		logger.Error("ttl error ",e.Error())
		return false,constants.RC_SYSTEM_ERR
	}
	if t <0 {
		setAndExpire(key,1, getTime())
	} else {

		if incrFlag {
			c,e := incr(key)
			if e != nil {
				logger.Error("incr error ",e.Error())
				return false,constants.RC_SYSTEM_ERR
			}
			if c > limit {
				return false,constants.RC_TOO_MANY_REQ
			}
		}else{
			c,e := rdsGet(key)
			if e != nil {
				logger.Error("incr error ",e.Error())
				return false,constants.RC_SYSTEM_ERR
			}
			if c > limit {
				return false,constants.RC_TOO_MANY_REQ
			}
		}

	}
	return true,constants.RC_OK
}


func CheckPrepareLimit(lvtUid int64,level int)(bool,constants.Error){
	key := DAILY_PREPARE_KEY_PROXY + utils.Int642Str(lvtUid)
	limit := getCFG(level)
	if limit == nil {
		return false,constants.RC_TOO_MANY_REQ
	}
	return checkLimit(key,limit.DailyPrepareAccess,true)
}

func CheckCommitLimit(lvtUid int64,level int)(bool,constants.Error){
	key := DAILY_COMMIT_KEY_PROXY + utils.Int642Str(lvtUid)
	limit := getCFG(level)
	if limit == nil {
		return false,constants.RC_TOO_MANY_REQ
	}
	return checkLimit(key,limit.DailyCommitAccess,false)
}

func checkTotalTransfer(lvtUid,amount int64,limit *config.TransferLimit)(bool,constants.Error){
	key := DAILY_TOTAL_TRANSFER_KEY_PROXY + utils.Int642Str(lvtUid)
	t,e :=  ttl(key)
	if e != nil {
		logger.Error("ttl error ",e.Error())
		return false,constants.RC_SYSTEM_ERR
	}
	if t <0 {
		setAndExpire(key,0, getTime())
	} else {
		total,err := rdsGet(key)
		if err != nil {
			logger.Error("redis get error ",err.Error())
			return false,constants.RC_SYSTEM_ERR
		}

		if (amount + int64(total)) > (limit.DailyAmountMax * LVT_CONV) {
			return false,constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
		}

	}
	return true,constants.RC_OK
}


func checkSingleAmount(amount int64,limit *config.TransferLimit)(bool,constants.Error){

	if amount > (limit.SingleAmountMax * LVT_CONV)  {
		return false,constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
	}
	if amount < (limit.SingleAmountMin * LVT_CONV)  {
		return false,constants.RC_TRANS_AMOUNT_TOO_LITTLE
	}
	return true,constants.RC_OK
}

func CheckAmount(lvtUid,amount int64,level int)(bool,constants.Error){
	limit := getCFG(level)
	if limit == nil {
		return false,constants.RC_TOO_MANY_REQ
	}
	if f,e := checkSingleAmount(amount,limit);!f {
		return false,e
	}
	return checkTotalTransfer(lvtUid,amount,limit)
}

func SetTotalTransfer(lvtUid,amount int64){
	commitKey := DAILY_COMMIT_KEY_PROXY + utils.Int642Str(lvtUid)
	totalKey := DAILY_TOTAL_TRANSFER_KEY_PROXY + utils.Int642Str(lvtUid)
	incr(commitKey)
	incrby(totalKey,amount)
}

func getTime()int{
	ts := utils.GetTimestamp13()
	start := utils.GetDayStart(ts)
	re := DAY_TS - (ts-utils.GetTimestamp13ByTime(start))
	return int(re/TS)
}



func GetTransLevel(uid int64)int{
	key := USER_TRANS_KEY_PROXY + utils.Int642Str(uid)
	t,err := ttl(key)
	if err != nil {
		return 0
	}
	var userTransLevel = 0
	var e error = nil
	if t <0 {
		userTransLevel = GetUserAssetTranslevelByUid(uid)
		setAndExpire(key,userTransLevel, DAY_30)
	} else {
		userTransLevel,e = rdsGet(key)
		if e != nil {
			logger.Error("get redis error")
			return 0
		}
	}
	return userTransLevel
}