package common

import (
	"utils"
	"utils/logger"
)

const (
	DEVICE_UNBIND_PROXY  = "de:ub:"
	DEVICE_LOCK_PROXY    = "de:lc:"
	DEVICE_UNBIND_EXPIRE = 24 * 3600
	DEVICE_LOCK_EXPIRE   = 50 * 5
)

func SetUnbindLimt(uid int64,mid int) {
	key := DEVICE_UNBIND_PROXY + utils.Int642Str(uid)+":"+utils.Int2Str(mid)
	setAndExpire(key, 1, DEVICE_UNBIND_EXPIRE)
}
func CheckUnbindLimit(uid int64,mid int) bool {
	key := DEVICE_UNBIND_PROXY + utils.Int642Str(uid)+":"+utils.Int2Str(mid)
	i, e := ttl(key)
	if e != nil {
		logger.Error("ttl redis error", e.Error())
		return false
	}
	return i > 0
}

func devicelock(key string)int64{
	ts := utils.GetTimestamp13()
	f,err := setnx(key,ts)
	if err != nil || f != 1 {
		return 0
	}
	return ts
}

func DeviceUserLock(uid int64) int64{
	key := DEVICE_LOCK_PROXY + utils.Int642Str(uid)
	ts := devicelock(key)
	if ts > 0 {
		rdsExpire(key,DEVICE_LOCK_EXPIRE)
	}
	return ts
}
func DeviceLock(appid int,did string) int64{
	key := DEVICE_LOCK_PROXY + utils.Int2Str(appid) + ":" + did
	ts := devicelock(key)
	if ts > 0 {
		rdsExpire(key,DEVICE_LOCK_EXPIRE)
	}
	return ts
}

func CheckDeviceLockDid(did string) bool {
	key := DEVICE_LOCK_PROXY + did
	i, e := ttl(key)
	if e != nil {
		logger.Error("ttl redis error", e.Error())
		return false
	}
	return i > 0
}

func DeviceUnLockUid(uid int64,ts int64) {
	key := DEVICE_LOCK_PROXY + utils.Int642Str(uid)
	t,e := rdsGet64(key)
	if e == nil && ts == t {
		rdsDel(key)
	}
}
func DeviceUnLockDid(appid int,did string,ts int64) {
	key := DEVICE_LOCK_PROXY + utils.Int2Str(appid) + ":" + did
	t,e := rdsGet64(key)
	if e == nil && ts == t {
		rdsDel(key)
	}
}
