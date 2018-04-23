package common

import (
	"utils/logger"
	"utils"
	"utils/config"
)


const (
	LOCK_ASSET_MONTH = 3
	DEF_SCORE = 70
)


func UserUpgrade(uid string)(bool,int){

	account,err := GetAccountByUID(uid)
	if err != nil {
		logger.Error("query account error",err.Error())
		return false,0
	}
	switch account.Level {
	case 0:
		return upZero(account)
	case 1:
		return upOne(account)
	case 2:
		return upTwo(account)
	case 3:
		return upThree(account)
	}

	return false,0
}


/**
	0
	set nickname
	set tx_pwd
	miner_days>3
	bind phone
 */
func upZero(acc *Account)(bool,int){
	// check base info
	if len(acc.Nickname)>0 && len(acc.PaymentPassword) > 0 && len(acc.Phone) >0 {
		// check miner days
		if QueryCountMinerByTs(acc.UID) > 3 && CheckCreditScore(acc.UID,DEF_SCORE){
			// set level up
			level := 1
			err := SetUserLevel(acc.UID,level)
			if err == nil {
				SetTransUserLevel(acc.UID,level)
				return true,level
			}
		}

	}
	return false,acc.Level
}
/**
	1
	miner_days>7
	lock_asset:month>=3,value>=1k
	bind wx(86)
 */
func upOne(acc *Account)(bool,int){
	// check miner days and bind wxid
	if CheckCreditScore(acc.UID,DEF_SCORE) && CheckBindWx(acc.UID) && QueryCountMinerByTs(acc.UID) > 7 {
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(1000)
		if v := QuerySumLockAsset(acc.UID,LOCK_ASSET_MONTH); v >= lvt {
			// set level up
			level := 2
			err := SetUserLevel(acc.UID,level)
			if err == nil {
				SetTransUserLevel(acc.UID,level)
				return true,level
			}
		}

	}
	return false,acc.Level
}
/**
	2
	miner_days>30
	lock_asset:month>=3,value>=5w
 */
func upTwo(acc *Account)(bool,int){
	// check miner days
	if QueryCountMinerByTs(acc.UID) > 30 && CheckCreditScore(acc.UID,DEF_SCORE){
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(50000)
		if v := QuerySumLockAsset(acc.UID,LOCK_ASSET_MONTH); v >= lvt {
			// set level up
			level := 3
			err := SetUserLevel(acc.UID,level)
			if err == nil {
				SetTransUserLevel(acc.UID,level)
				return true,level
			}
		}

	}
	return false,acc.Level
}
/**
	3
	miner_days>100
	lock_asset:month>=3,value>=20w
 */
func upThree(acc *Account)(bool,int){
	// check miner days
	if QueryCountMinerByTs(acc.UID) > 100 && CheckCreditScore(acc.UID,DEF_SCORE){
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(200000)
		if v := QuerySumLockAsset(acc.UID,LOCK_ASSET_MONTH); v >= lvt {
			// set level up
			level := 4
			err := SetUserLevel(acc.UID,level)
			if err == nil {
				SetTransUserLevel(acc.UID,level)
				return true,level
			}
		}

	}
	return false,acc.Level
}


func CanBeTo(uid int64)bool{
	if config.GetConfig().CautionMoneyIdsExist(uid) {
		return true
	}
	return getUserLimit(uid).TransferTo()
}


func CanLockAsset(uid int64)bool{
	return getUserLimit(uid).LockAsset()
}

func CheckCreditScore(uid int64,score int)bool{
	creditScore := GetUserCreditScore(uid)
	return creditScore>=score
}

func getUserLimit(uid int64)*config.UserLevelLimit{
	level := GetTransUserLevel(uid)
	limit := config.GetLimitByLevel(level)
	logger.Info("user level",level,"limit",utils.ToJSON(*limit))
	return limit
}