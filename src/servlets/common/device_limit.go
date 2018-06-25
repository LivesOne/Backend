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

func DeviceLockUid(uid int64) {
	key := DEVICE_LOCK_PROXY + utils.Int642Str(uid)
	setAndExpire(key, 1, DEVICE_LOCK_EXPIRE)
}
func DeviceLockDid(did string) {
	key := DEVICE_LOCK_PROXY + did
	setAndExpire(key, 1, DEVICE_LOCK_EXPIRE)
}

func CheckDeviceLockUid(uid int64) bool {
	key := DEVICE_LOCK_PROXY + utils.Int642Str(uid)
	i, e := ttl(key)
	if e != nil {
		logger.Error("ttl redis error", e.Error())
		return false
	}
	return i > 0
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

func DeviceUnLockUid(uid int64) {
	key := DEVICE_LOCK_PROXY + utils.Int642Str(uid)
	rdsDel(key)
}
func DeviceUnLockDid(did string) {
	key := DEVICE_LOCK_PROXY + did
	rdsDel(key)
}
