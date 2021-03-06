package common

import (
	"servlets/constants"
	"strconv"
	"utils"
	"utils/config"
	"utils/logger"
)

const (
	DAILY_PREPARE_KEY_PROXY         = "tl:dp:"
	DAILY_COMMIT_KEY_PROXY          = "tl:dc:"
	DAILY_TOTAL_TRANSFER_KEY_PROXY  = "tl:dt:"
	DAILY_TRANS_LVTC_KEY_PROXY      = "tl:dt:lvtc:"
	DAILY_TRANS_ETH_KEY_PROXY       = "tl:dt:eth:"
	DAILY_TRANS_EOS_KEY_PROXY       = "tl:dt:eos:"
	DAILY_TRANS_BTC_KEY_PROXY       = "tl:dt:btc:"
	TRANS_SINGLE_MIN_LVTC_KEY_PROXY = "tl:tsm:lvtc"
	TRANS_SINGLE_MIN_ETH_KEY_PROXY  = "tl:tsm:eth"
	TRANS_SINGLE_MIN_EOS_KEY_PROXY  = "tl:tsm:eos"
	TRANS_SINGLE_MIN_BTC_KEY_PROXY  = "tl:tsm:btc"
	TRANS_DAILY_MAX_LVTC_KEY_PROXY = "tl:tdm:lvtc"
	TRANS_DAILY_MAX_ETH_KEY_PROXY  = "tl:tdm:eth"
	TRANS_DAILY_MAX_EOS_KEY_PROXY  = "tl:tdm:eos"
	TRANS_DAILY_MAX_BTC_KEY_PROXY  = "tl:tdm:btc"
	USER_TRANS_KEY_PROXY            = "tx:uid:"
	USER_LEVEL_KEY_PROXY            = "tx:ul:"
	TS                              = 1000
	DAY_S                           = 24 * 3600
	DAY_30                          = DAY_S * 30
	DAY_TS                          = DAY_S * TS
	LVT_CONV                        = 100000000
)

var cfg map[int]config.TransferLimit

func getCFG(userLevel int) *config.TransferLimit {
	if cfg == nil {
		cfg = config.GetConfig().TransferLimit
	}
	if lim, ok := cfg[userLevel]; ok {
		return &lim
	}
	return nil
}

func checkLimit(key string, limit int, incrFlag bool) (bool, constants.Error) {
	t, e := ttl(key)
	if e != nil {
		logger.Error("ttl error ", e.Error())
		return false, constants.RC_SYSTEM_ERR
	}
	if t < 0 {
		setAndExpire(key, 1, getTime())
	} else {
		var c int
		var e error
		if incrFlag {
			c, e = incr(key)
		} else {
			c, e = rdsGet(key)
		}
		if e != nil {
			logger.Error("incr error ", e.Error())
			return false, constants.RC_SYSTEM_ERR
		}
		if limit > -1 && c > limit {
			return false, constants.RC_TOO_MANY_REQ
		}

	}
	return true, constants.RC_OK
}

func CheckPrepareLimit(lvtUid int64, level int) (bool, constants.Error) {
	key := DAILY_PREPARE_KEY_PROXY + utils.Int642Str(lvtUid)
	var limit int
	//交易员等级为0的话，去校验用户等级
	if level == 0 {
		userLevel := GetTransUserLevel(lvtUid)
		limitConfig := config.GetLimitByLevel(userLevel)
		limit = limitConfig.DailyPrepareAccess()
	} else {
		limitConfig := getCFG(level)
		if limitConfig == nil {
			return false, constants.RC_TOO_MANY_REQ
		}
		limit = limitConfig.DailyPrepareAccess
	}
	return checkLimit(key, limit, true)
}

func CheckCommitLimit(lvtUid int64, level int) (bool, constants.Error) {
	key := DAILY_COMMIT_KEY_PROXY + utils.Int642Str(lvtUid)
	var limit int
	//交易员等级为0的话，去校验用户等级
	if level == 0 {
		userLevel := GetTransUserLevel(lvtUid)
		limitConfig := config.GetLimitByLevel(userLevel)
		limit = limitConfig.DailyCommitAccess()
	} else {
		limitConfig := getCFG(level)
		if limitConfig == nil {
			return false, constants.RC_TOO_MANY_REQ
		}
		limit = limitConfig.DailyCommitAccess
	}
	return checkLimit(key, limit, false)
}

func checkTransAmount(key, currency string, maxAmount bool) (int64, constants.Error) {
	t, e := ttl(key)
	if e != nil {
		logger.Error("ttl error ", e.Error())
		return 0, constants.RC_SYSTEM_ERR
	}
	var limit int64
	if t < 0 {
		var amount float64
		var err error
		if maxAmount {
			amount, err = QueryTransDailyAmountMax(currency)
		} else {
			amount, err = QueryTransSingleAmountMin(currency)
		}
		if err != nil {
			return 0, constants.RC_SYSTEM_ERR
		}
		limitStr := strconv.FormatFloat(amount, 'f', -1, 64)
		var limitInt int64
		if currency == constants.TRADE_CURRENCY_EOS {
			limitInt = utils.FloatStrToEOSint(limitStr)
		} else {
			limitInt = utils.FloatStrToLVTint(limitStr)
		}
		setAndExpire64(key, limitInt, getTime())
		limit = limitInt
	} else {
		c, e := rdsGet64(key)
		if e != nil {
			logger.Error("redis get:", key, " error ", e.Error())
			return 0, constants.RC_SYSTEM_ERR
		}
		limit = c
	}
	return limit, constants.RC_OK
}

func CheckSingleTransAmount(currency string, amount int64) constants.Error {
	var key string
	switch currency {
	case constants.TRADE_CURRENCY_LVTC:
		key = TRANS_SINGLE_MIN_LVTC_KEY_PROXY
	case constants.TRADE_CURRENCY_ETH:
		key = TRANS_SINGLE_MIN_ETH_KEY_PROXY
	case constants.TRADE_CURRENCY_EOS:
		key = TRANS_SINGLE_MIN_EOS_KEY_PROXY
	case constants.TRADE_CURRENCY_BTC:
		key = TRANS_SINGLE_MIN_BTC_KEY_PROXY
	default:
		return constants.RC_INVALID_CURRENCY
	}
	// 获取单笔转账限额
	limit, err := checkTransAmount(key, currency, false)
	if err != constants.RC_OK {
		return err
	}

	if limit > -1 && amount < limit {
		return constants.RC_TRANS_AMOUNT_TOO_LITTLE
	}
	return constants.RC_OK
}

func CheckDailyTransAmount(uid int64, currency string, amount int64) (bool, constants.Error) {
	var key, keyMaxLimit string
	switch currency {
	case constants.TRADE_CURRENCY_LVTC:
		key = DAILY_TRANS_LVTC_KEY_PROXY
		keyMaxLimit = TRANS_DAILY_MAX_LVTC_KEY_PROXY
	case constants.TRADE_CURRENCY_ETH:
		key = DAILY_TRANS_ETH_KEY_PROXY
		keyMaxLimit = TRANS_DAILY_MAX_ETH_KEY_PROXY
	case constants.TRADE_CURRENCY_EOS:
		key = DAILY_TRANS_EOS_KEY_PROXY
		keyMaxLimit = TRANS_DAILY_MAX_EOS_KEY_PROXY
	case constants.TRADE_CURRENCY_BTC:
		key = DAILY_TRANS_BTC_KEY_PROXY
		keyMaxLimit = TRANS_DAILY_MAX_BTC_KEY_PROXY
	default:
		return false, constants.RC_INVALID_CURRENCY
	}
	key += utils.Int642Str(uid)
	// 获取日转账限额
	limit, err := checkTransAmount(keyMaxLimit, currency, true)
	if err != constants.RC_OK {
		return false, err
	}

	t, e := ttl(key)
	if e != nil {
		logger.Error("ttl error ", e.Error())
		return false, constants.RC_SYSTEM_ERR
	}
	var curAmount int64
	if t < 0 {
		setAndExpire(key, 0, getTime())
	} else {
		c, e := rdsGet64(key)
		if e != nil {
			logger.Error("redis get:", key, " error ", e.Error())
			return false, constants.RC_SYSTEM_ERR
		}
		curAmount = c
	}
	if limit > -1 && (curAmount+amount) > limit {
		return false, constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
	}
	return true, constants.RC_OK
}

func SetDailyTransAmount(uid int64, currency string, amount int64) (bool, constants.Error) {
	var key string
	switch currency {
	case constants.TRADE_CURRENCY_LVTC:
		key = DAILY_TRANS_LVTC_KEY_PROXY
	case constants.TRADE_CURRENCY_ETH:
		key = DAILY_TRANS_ETH_KEY_PROXY
	case constants.TRADE_CURRENCY_EOS:
		key = DAILY_TRANS_EOS_KEY_PROXY
	case constants.TRADE_CURRENCY_BTC:
		key = DAILY_TRANS_BTC_KEY_PROXY
	default:
		return false, constants.RC_INVALID_CURRENCY
	}
	key += utils.Int642Str(uid)
	_, err := incrby(key, amount)
	if err != nil {
		return false, constants.RC_SYSTEM_ERR
	}
	return true, constants.RC_OK
}

func checkTotalTransfer(lvtUid, amount int64, limit *config.TransferLimit) (bool, constants.Error) {
	key := DAILY_TOTAL_TRANSFER_KEY_PROXY + utils.Int642Str(lvtUid)
	t, e := ttl(key)
	if e != nil {
		logger.Error("ttl error ", e.Error())
		return false, constants.RC_SYSTEM_ERR
	}
	var total int
	if t < 0 {
		setAndExpire(key, 0, getTime())
	} else {
		c, err := rdsGet(key)
		if err != nil {
			logger.Error("redis get error ", err.Error())
			return false, constants.RC_SYSTEM_ERR
		}
		total = c
	}
	if (limit.DailyAmountMax > -1) && (amount+int64(total)) > (limit.DailyAmountMax*LVT_CONV) {
		return false, constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
	}
	return true, constants.RC_OK
}

func checkTotalTransferByUserLevel(lvtUid, amount int64, limit *config.UserLevelLimit) (bool, constants.Error) {
	key := DAILY_TOTAL_TRANSFER_KEY_PROXY + utils.Int642Str(lvtUid)
	t, e := ttl(key)
	if e != nil {
		logger.Error("ttl error ", e.Error())
		return false, constants.RC_SYSTEM_ERR
	}
	var total int
	if t < 0 {
		setAndExpire(key, 0, getTime())
	} else {
		c, err := rdsGet(key)
		if err != nil {
			logger.Error("redis get error ", err.Error())
			return false, constants.RC_SYSTEM_ERR
		}
		total = c
	}
	if (limit.DailyAmountMax() > -1) && (amount+int64(total)) > (limit.DailyAmountMax()*LVT_CONV) {
		return false, constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
	}
	return true, constants.RC_OK
}

func checkSingleAmount(amount int64, limit *config.TransferLimit) (bool, constants.Error) {

	if limit.SingleAmountMax > -1 && amount > (limit.SingleAmountMax*LVT_CONV) {
		return false, constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
	}
	if limit.SingleAmountMin > -1 && amount < (limit.SingleAmountMin*LVT_CONV) {
		return false, constants.RC_TRANS_AMOUNT_TOO_LITTLE
	}
	return true, constants.RC_OK
}

func checkSingleAmountByUserLevel(amount int64, limit *config.UserLevelLimit) (bool, constants.Error) {

	if limit.SingleAmountMax() > -1 && amount > (limit.SingleAmountMax()*LVT_CONV) {
		return false, constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
	}
	if limit.SingleAmountMin() > -1 && amount < (limit.SingleAmountMin()*LVT_CONV) {
		return false, constants.RC_TRANS_AMOUNT_TOO_LITTLE
	}
	return true, constants.RC_OK
}

func CheckAmount(lvtUid, amount int64, level int) (bool, constants.Error) {
	//交易员等级为0的话，去校验用户等级
	if level == 0 {
		userLevel := GetTransUserLevel(lvtUid)
		limit := config.GetLimitByLevel(userLevel)
		if f, e := checkSingleAmountByUserLevel(amount, limit); !f {
			return false, e
		}
		return checkTotalTransferByUserLevel(lvtUid, amount, limit)
	} else {
		limit := getCFG(level)
		if limit == nil {
			return false, constants.RC_TOO_MANY_REQ
		}
		if f, e := checkSingleAmount(amount, limit); !f {
			return false, e
		}
		return checkTotalTransfer(lvtUid, amount, limit)
	}
}

func SetTotalTransfer(lvtUid, amount int64) {
	commitKey := DAILY_COMMIT_KEY_PROXY + utils.Int642Str(lvtUid)
	totalKey := DAILY_TOTAL_TRANSFER_KEY_PROXY + utils.Int642Str(lvtUid)
	incr(commitKey)
	incrby(totalKey, amount)
}

func getTime() int {
	ts := utils.GetTimestamp13()
	start := utils.GetDayStart(ts)
	re := DAY_TS - (ts - utils.GetTimestamp13ByTime(start))
	return int(re / TS)
}

func GetTransLevel(uid int64) int {
	key := USER_TRANS_KEY_PROXY + utils.Int642Str(uid)
	t, err := ttl(key)
	if err != nil {
		return 0
	}
	var userTransLevel = 0
	var e error = nil
	if t < 0 {
		userTransLevel = GetUserAssetTranslevelByUid(uid)
		setAndExpire(key, userTransLevel, DAY_30)
	} else {
		userTransLevel, e = rdsGet(key)
		if e != nil {
			logger.Error("get redis error")
			return 0
		}
	}
	return userTransLevel
}

func GetTransUserLevel(uid int64) int {
	key := USER_LEVEL_KEY_PROXY + utils.Int642Str(uid)
	//t, err := ttl(key)
	//logger.Info("ttl key", key, "expire ", t)
	//if err != nil {
	//	return 0
	//}
	//var userLevel = 0
	//var e error = nil
	//if t < 0 {
	//	userLevel = GetUserLevel(uid)
	//	logger.Info("key", key, "t ", t, userLevel)
	//	if userLevel > -1 {
	//		setAndExpire(key, userLevel, DAY_30)
	//	}
	//} else {
	//	userLevel, e = rdsGet(key)
	//	logger.Info("rdsGet key", key, "value ", userLevel)
	//	if e != nil {
	//		logger.Error("get redis error")
	//		return 0
	//	}
	//}
	//return userLevel
	userLevel := GetUserLevel(uid)
	logger.Info("key", key, userLevel)
	if userLevel > -1 {
		setAndExpire(key, userLevel, DAY_30)
	}
	return userLevel
}

func SetTransUserLevel(uid int64, userLevel int) {
	key := USER_LEVEL_KEY_PROXY + utils.Int642Str(uid)
	setAndExpire(key, userLevel, DAY_30)
}
