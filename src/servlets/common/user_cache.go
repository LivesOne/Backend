package common

import (
	"errors"
	"utils"
	"utils/logger"
)

const (
	USER_CACHE_REDIS_KEY_PROXY                = "cache:user:info:"
	INFO_OK                                   = "OK"
	HSET_OK                                   = 1
	USER_CACHE_EXPIRE_30DAY                   = 3600 * 24 * 30
	USER_CACHE_REDIS_FIELD_NAME_UID           = "uid"
	USER_CACHE_REDIS_FIELD_NAME_NICKNAME      = "nickname"
	USER_CACHE_REDIS_FIELD_NAME_EMAIL         = "email"
	USER_CACHE_REDIS_FIELD_NAME_COUNTRY       = "country"
	USER_CACHE_REDIS_FIELD_NAME_PHONE         = "phone"
	USER_CACHE_REDIS_FIELD_NAME_LEVEL         = "level"
	USER_CACHE_REDIS_FIELD_NAME_CREDITR_SCORE = "credit_score"
	USER_CACHE_REDIS_FIELD_NAME_AVATAR_URL    = "avatar_url"
	USER_CACHE_REDIS_FIELD_NAME_ACTIVE_DAYS   = "active_days"
)

func GetCacheUser(uid int64) (map[string]string, error) {
	if e := ttlAndInit(uid); e != nil {
		return nil, e
	}
	return hgetall(buildKey(uid))
}

func GetCacheUserField(uid int64, fieldName string) (string, error) {
	if e := ttlAndInit(uid); e != nil {
		return "", e
	}
	return hget(buildKey(uid), fieldName)
}

func SetCacheUserField(uid int64, fieldName, fieldValue string) bool {
	if e := ttlAndInit(uid); e != nil {
		return false
	}
	info, err := hset(buildKey(uid), fieldName, fieldValue)
	if err != nil || info != HSET_OK {
		return false
	}
	return true
}

func ttlAndInit(uid int64) error {
	key := buildKey(uid)
	c, e := ttl(key)
	if e != nil {
		logger.Error("redis ttl error", e.Error())
		return e
	}
	if c < -1 {
		u, e := QueryCacheUser(uid)
		if e != nil {
			return e
		}
		if len(u) == 0 {
			return errors.New("can not find user by uid " + utils.Int642Str(uid))
		}
		info, err := hmset(key, u)
		if err != nil {
			return err
		}
		if info != INFO_OK {
			return errors.New("can not save user to redis by uid " + utils.Int642Str(uid))
		}
		rdsExpire(key, USER_CACHE_EXPIRE_30DAY)
	}
	return nil
}

func buildKey(uid int64) string {
	return USER_CACHE_REDIS_KEY_PROXY + utils.Int642Str(uid)
}
