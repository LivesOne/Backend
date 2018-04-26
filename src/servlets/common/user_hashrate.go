package common

import (
	"utils"
	"utils/config"
	"servlets/constants"
)

const (
	USER_HASHRATE_TYPE_LOCK_ASSET = 1
	USER_HASHRATE_TYPE_BIND_WX = 2
	USER_HASHRATE_TYPE_BIND_TG = 3

)




func AddBindActiveHashRateByWX(uid int64){
	ts := utils.GetTimestamp13()
	limit := config.GetBindActive()
	if checkActiveTime(uid,ts,limit) {
		end := (int64(limit.HashRateActiveMonth) * constants.ASSET_LOCK_MONTH_TIMESTAMP) + ts
		updHashRate(uid,limit.BindWXActiveHashRate,USER_HASHRATE_TYPE_BIND_WX,ts,end,nil)
	}

}

func AddBindActiveHashRateByTG(uid int64){
	ts := utils.GetTimestamp13()
	limit := config.GetBindActive()
	if checkActiveTime(uid,ts,limit) {
		end := (int64(limit.HashRateActiveMonth) * constants.ASSET_LOCK_MONTH_TIMESTAMP) + ts
		updHashRate(uid,limit.BindWXActiveHashRate,USER_HASHRATE_TYPE_BIND_TG,ts,end,nil)
	}

}


func checkActiveTime(uid,ts int64,limit *config.BindActive)bool{
	//活动期间内直接加入算力
	if ts >= limit.BindTimeActiveStart && ts <= limit.BindTimeActiveEnd {
		return true
	} else {
		//数据库中存储到秒
		userRegisterTs := GetUserRegisterTime(uid) * 1000

		if ts - userRegisterTs <= limit.RegisterTimeActive {
			return true
		}
	}
	return false
}