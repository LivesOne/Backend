package common

import (
	"servlets/constants"
	"utils"
	"utils/logger"
)

const (
	USE_CARD_KEY_PROXY = "uc:dl:"
)

func CheckUserCardLimit(uid int64) (bool, constants.Error) {
	key := USE_CARD_KEY_PROXY + utils.Int642Str(uid)
	t, e := ttl(key)
	if e != nil {
		logger.Error("ttl error ", e.Error())
		return false, constants.RC_SYSTEM_ERR
	}
	if t > 0 {
		c, e := rdsGet(key)
		if e != nil {
			logger.Error("incr error ", e.Error())
			return false, constants.RC_SYSTEM_ERR
		}
		if c >= 5 {
			return false, constants.RC_TOO_MANY_REQ
		}
	}
	return true, constants.RC_OK
}

func AddUseCardLimit(uid int64) {
	key := USE_CARD_KEY_PROXY + utils.Int642Str(uid)
	_, e := incr(key)
	if e != nil {
		logger.Error("incr error ", e.Error())
	}
}
