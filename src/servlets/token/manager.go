package token

import (
	"servlets/common"
	"servlets/constants"
	"utils"
	"utils/config"
	"utils/logger"

	"github.com/thanhpk/randstr"
)

var gDB RedisDB

func Init() {
	gDB = RedisDB{}
	logger.Debug(config.GetConfig().RedisAddr)
	common.Init_redis(map[string]string{
		"addr": config.GetConfig().RedisAddr,
		"auth": config.GetConfig().RedisAuth,
	})
}

func New(uid, key string, expire int64) (newtoken string, err int) {
	ret := constants.ERR_INT_TK_DUPLICATE
	token := ""
	for ret == constants.ERR_INT_TK_DUPLICATE {
		token = randstr.String(32)
		hash := utils.Sha256(token)
		ret = gDB.Insert(hash, uid, key, token, expire)
	}
	if ret != constants.ERR_INT_OK {
		return "", ret
	} else {
		return token, ret
	}
}

func Update(hash, key string, expire int64) (err int) {
	return gDB.Update(hash, key, expire)
}

func Del(hash string) (err int) {
	return gDB.Delete(hash)
}

func GetUID(hash string) (uid string, err int) {
	return gDB.GetUID(hash)
}

func GetKey(hash string) (key string, err int) {
	return gDB.GetKey(hash)
}

func GetToken(hash string) (token string, err int) {
	return gDB.GetToken(hash)
}

func GetAll(hash string) (uid, key, token string, err int) {
	return gDB.GetAll(hash)
}
