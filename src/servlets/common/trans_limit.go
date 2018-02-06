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
	TS = 1000
	DAY_TS = 24*3600*TS
	LVT_CONV = 100000000
)

var cfg *config.TransferLimit

func getCFG()*config.TransferLimit{
	if cfg == nil {
		cfg = &config.GetConfig().TransferLimit
	}
	return cfg
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
	_,err := rdsDo("SET",key,value)
	if err != nil {
		return err
	}
	_,err = rdsDo("EXPIRE",key,expire)
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


func CheckPrepareLimit(lvtUid int64)(bool,constants.Error){
	key := DAILY_PREPARE_KEY_PROXY + utils.Int642Str(lvtUid)
	return checkLimit(key,getCFG().DailyPrepareAccecss,true)
}

func CheckCommitLimit(lvtUid int64)(bool,constants.Error){
	key := DAILY_COMMIT_KEY_PROXY + utils.Int642Str(lvtUid)
	return checkLimit(key,getCFG().DailyCommitAccess,false)
}

func checkTotalTransfer(lvtUid,amount int64)(bool,constants.Error){
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
		if (amount + int64(total)) > (getCFG().DailyAmountMax * LVT_CONV) {
			return false,constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
		}

	}
	return true,constants.RC_OK
}


func checkSingleAmount(amount int64)(bool,constants.Error){
	if amount > (getCFG().SingleAmountMax * LVT_CONV)  {
		return false,constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
	}
	if amount < (getCFG().SingleAmountMin * LVT_CONV)  {
		return false,constants.RC_TRANS_AMOUNT_TOO_LITTLE
	}
	return true,constants.RC_OK
}

func CheckAmount(lvtUid,amount int64)(bool,constants.Error){
	if f,e := checkSingleAmount(amount);!f {
		return false,e
	}
	return checkTotalTransfer(lvtUid,amount)
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