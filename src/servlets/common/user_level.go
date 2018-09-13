package common

import (
	"utils"
	"utils/config"
	"utils/logger"
)

const (
	LOCK_ASSET_MONTH = 3
	DEF_SCORE        = 70
	DEF_LEVEL        = 2
)

func UserUpgrade(uid string) (bool, int) {

	account, err := GetAccountByUID(uid)
	if err != nil {
		logger.Error("query account error", err.Error())
		return false, 0
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

	return false, 0
}

/**
0
set nickname
set tx_pwd
miner_days>3
bind phone
*/
func upZero(acc *Account) (bool, int) {
	// check base info
	if len(acc.Nickname) > 0 && len(acc.PaymentPassword) > 0 && len(acc.Phone) > 0 {
		// check miner days
		if QueryUserActiveDaysByCache(acc.UID) >= 3 && CheckCreditScore(acc.UID, DEF_SCORE) {
			// set level up
			level := 1
			err := SetUserLevel(acc.UID, level)
			if err == nil {
				SetTransUserLevel(acc.UID, level)
				return true, level
			}
		}

	}
	return false, acc.Level
}

/**
1
miner_days>7
lock_asset:month>=3,value>=1k
bind wx(86)
*/
func upOne(acc *Account) (bool, int) {
	// check miner days and bind wxid
	lvtcScale := int64(config.GetConfig().LvtcHashrateScale)
	if CheckCreditScore(acc.UID, DEF_SCORE) && CheckBindWx(acc.UID) &&
		QueryUserActiveDaysByCache(acc.UID) >= 7 && lvtcScale > 0 {
		//check asset lock month and value
		lvtc := utils.CONV_LVT * int64(1000) / lvtcScale
		if v := QuerySumLockAssetLvtc(acc.UID, LOCK_ASSET_MONTH, CURRENCY_LVTC); v >= lvtc {
			// set level up
			level := 2
			err := SetUserLevel(acc.UID, level)
			if err == nil {
				SetTransUserLevel(acc.UID, level)
				return true, level
			}
		}

	}
	return false, acc.Level
}

/**
2
miner_days>30
lock_asset:month>=3,value>=5w
*/
func upTwo(acc *Account) (bool, int) {
	// check miner days
	lvtcScale := int64(config.GetConfig().LvtcHashrateScale)
	if QueryUserActiveDaysByCache(acc.UID) >= 30 && CheckCreditScore(acc.UID, DEF_SCORE) && lvtcScale > 0 {
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(50000) / lvtcScale
		if v := QuerySumLockAssetLvtc(acc.UID, LOCK_ASSET_MONTH, CURRENCY_LVTC); v >= lvt {
			// set level up
			level := 3
			err := SetUserLevel(acc.UID, level)
			if err == nil {
				SetTransUserLevel(acc.UID, level)
				return true, level
			}
		}

	}
	return false, acc.Level
}

/**
3
miner_days>90
lock_asset:month>=3,value>=20w
*/
func upThree(acc *Account) (bool, int) {
	// check miner days
	lvtcScale := int64(config.GetConfig().LvtcHashrateScale)
	if QueryUserActiveDaysByCache(acc.UID) >= 90 && CheckCreditScore(acc.UID, DEF_SCORE) && lvtcScale > 0 {
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(200000) / lvtcScale
		if v := QuerySumLockAssetLvtc(acc.UID, LOCK_ASSET_MONTH, CURRENCY_LVTC); v >= lvt {
			// set level up
			level := 4
			err := SetUserLevel(acc.UID, level)
			if err == nil {
				SetTransUserLevel(acc.UID, level)
				return true, level
			}
		}

	}
	return false, acc.Level
}

func CanBeTo(uid int64) bool {
	return getUserLimit(uid).TransferTo()
}

func CanLockAsset(uid int64) bool {
	return getUserLimit(uid).LockAsset()
}

func CheckCreditScore(uid int64, score int) bool {
	creditScore := GetUserCreditScore(uid)
	return creditScore >= score
}

func getUserLimit(uid int64) *config.UserLevelLimit {
	level := GetTransUserLevel(uid)
	limit := config.GetLimitByLevel(level)
	logger.Info("user level", level, "limit", utils.ToJSON(*limit))
	return limit
}

func CheckUserLevel(uid int64, level int) bool {
	userLevel := GetUserLevel(uid)
	return userLevel >= level
}
