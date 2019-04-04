package common

import (
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"servlets/rpc"
	"strings"
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

	account, err := rpc.GetUserInfo(utils.Str2Int64(uid))
	if err != nil {
		logger.Error("query account error", err.Error())
		return false, 0
	}
	if account.Result != microuser.ResCode_OK {
		logger.Error("query account error", account.Result.String())
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
func upZero(acc *microuser.GetUserAllInfoRes) (bool, int) {
	// check base info
	if len(acc.Nickname) > 0 && len(acc.PaymentPassword) > 0 && len(acc.Phone) > 0 {
		// check miner days
		if acc.ActiveDays >= 3 && acc.CreditScore >= DEF_SCORE {
			// set level up
			level := "1"
			f, err := rpc.SetUserField(acc.Uid, microuser.UserField_LEVEL, level)
			if err == nil && f {
				return true, utils.Str2Int(level)
			}
		}

	}
	return false, int(acc.Level)
}

/**
1
miner_days>7
lock_asset:month>=3,value>=1k
bind wx(86)
*/
func upOne(acc *microuser.GetUserAllInfoRes) (bool, int) {
	// check miner days and bind wxid
	lvtcScale := int64(config.GetConfig().LvtcHashrateScale)
	if acc.ActiveDays >= 7 && acc.CreditScore >= DEF_SCORE && CheckBindWx(acc.Wx,acc.Country) && lvtcScale > 0 {
		//check asset lock month and value
		lvtc := utils.CONV_LVT * int64(1000) / lvtcScale
		if v := QuerySumLockAssetLvtc(acc.Uid, LOCK_ASSET_MONTH, CURRENCY_LVTC); v >= lvtc {
			// set level up
			level := "2"
			f, err := rpc.SetUserField(acc.Uid, microuser.UserField_LEVEL, level)
			if err == nil && f {
				return true, utils.Str2Int(level)
			}
		}

	}
	return false, int(acc.Level)
}

/**
2
miner_days>30
lock_asset:month>=3,value>=5w
*/
func upTwo(acc *microuser.GetUserAllInfoRes) (bool, int) {
	// check miner days
	lvtcScale := int64(config.GetConfig().LvtcHashrateScale)
	if acc.ActiveDays >= 30 && acc.CreditScore >= DEF_SCORE && lvtcScale > 0 {
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(50000) / lvtcScale
		if v := QuerySumLockAssetLvtc(acc.Uid, LOCK_ASSET_MONTH, CURRENCY_LVTC); v >= lvt {
			// set level up
			level := "3"
			f, err := rpc.SetUserField(acc.Uid, microuser.UserField_LEVEL, level)
			if err == nil && f {
				return true, utils.Str2Int(level)
			}
		}

	}
	return false, int(acc.Level)
}

/**
3
miner_days>100
lock_asset:month>=3,value>=20w
*/
func upThree(acc *microuser.GetUserAllInfoRes) (bool, int) {
	// check miner days
	lvtcScale := int64(config.GetConfig().LvtcHashrateScale)
	if acc.ActiveDays >= 100 && acc.CreditScore >= DEF_SCORE && lvtcScale > 0 {
		//check asset lock month and value
		lvt := utils.CONV_LVT * int64(200000) / lvtcScale
		if v := QuerySumLockAssetLvtc(acc.Uid, LOCK_ASSET_MONTH, CURRENCY_LVTC); v >= lvt {
			// set level up
			level := "4"
			f, err := rpc.SetUserField(acc.Uid, microuser.UserField_LEVEL, level)
			if err == nil && f {
				return true, utils.Str2Int(level)
			}
		}

	}
	return false, int(acc.Level)
}

func CanBeTo(uid int64) bool {
	return getUserLimit(uid).TransferTo()
}

func CanLockAsset(uid int64) bool {
	return getUserLimit(uid).LockAsset()
}

func getUserLimit(uid int64) *config.UserLevelLimit {
	level := GetTransUserLevel(uid)
	limit := config.GetLimitByLevel(level)
	logger.Info("user level", level, "limit", utils.ToJSON(*limit))
	return limit
}

func CheckBindWx(wx string,area int64) bool {
	switch(area){
	case 86:
		fallthrough
	case 852:
		fallthrough
	case 853:
		fallthrough
	case 886:
		ids := strings.Split(wx, ",")
		if len(ids) != 2 {
			return false
		}
		if len(ids[0]) == 0 || len(ids[1]) == 0 {
			return false
		}
	}
	return true
}
