package token

import (
	"servlets/constants"
	"utils/config"

	"github.com/thanhpk/randstr"
)

var gDB Database

func Init() {
	gDB := &RedisDB{}
	gDB.Open(config.GetConfig().RedisAddr)
}

func New(uid, key string, expire int) (newtoken string, err int) {
	ret := constants.ERR_INT_TK_DUPLICATE
	token := ""
	for ret == constants.ERR_INT_TK_DUPLICATE {
		token = randstr.String(32)
		hash := token //hash method should be sha256
		ret = gDB.Insert(hash, uid, key, token, expire)
	}
	if ret != constants.ERR_INT_OK {
		return "", ret
	} else {
		return token, ret
	}
}

func Update(hash, key string, expire int) (err int) {
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
