package common

import (
	"utils/logger"
	"utils"
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
		if QueryCountMinerByTs(acc.UID) > 3 {
			// set level up
			err := SetUserLevel(acc.UID,1)
			if err == nil {
				return true,1
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
	if CheckBindWx(acc.UID) && QueryCountMinerByTs(acc.UID) > 7 {
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(1000)
		if m,v := QuerySumLockAsset(acc.UID);m >= 3 && v >= lvt {
			// set level up
			err := SetUserLevel(acc.UID,2)
			if err == nil {
				return true,2
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
	if QueryCountMinerByTs(acc.UID) > 30 {
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(50000)
		if m,v := QuerySumLockAsset(acc.UID);m >= 3 && v >= lvt {
			// set level up
			err := SetUserLevel(acc.UID,3)
			if err == nil {
				return true,3
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
	if QueryCountMinerByTs(acc.UID) > 100 {
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(200000)
		if m,v := QuerySumLockAsset(acc.UID);m >= 3 && v >= lvt {
			// set level up
			err := SetUserLevel(acc.UID,3)
			if err == nil {
				return true,4
			}
		}

	}
	return false,acc.Level
}