package token

import (
	"servlets/common"
	"servlets/constants"
	"time"

	"github.com/garyburd/redigo/redis"
)

type RedisDB struct {
}

func (r *RedisDB) Insert(hash, uid, key, token string, expire int64) int {
	conn := common.GetRedisConn()
	defer conn.Close()

	// _, err := conn.Do("WATCH", "tk:"+hash)
	// if err != nil {
	// 	return TK_ERR_DB
	// }

	exists, err := redis.Bool(conn.Do("EXISTS", "tk:"+hash))
	if err != nil {
		return constants.ERR_INT_TK_DB
	} else if exists {
		// conn.Do("UNWATCH")
		return constants.ERR_INT_TK_DUPLICATE
	}

	_, err = conn.Do("HMSET", "tk:"+hash, "uid", uid, "key", key, "token", token, "exp", expire+time.Now().Unix())
	if err != nil {
		return constants.ERR_INT_TK_DB
	} else {
		conn.Do("EXPIRE", "tk:"+hash, expire)
		return constants.ERR_INT_OK
	}
}

func (r *RedisDB) Update(hash, key string, expire int64) int {
	conn := common.GetRedisConn()
	defer conn.Close()

	// _, err := conn.Do("WATCH", "tk:"+hash)
	// if err != nil {
	// 	return TK_ERR_DB
	// }

	exists, err := redis.Bool(conn.Do("EXISTS", "tk:"+hash))
	if err != nil {
		return constants.ERR_INT_TK_DB
	} else if !exists {
		// conn.Do("UNWATCH")
		return constants.ERR_INT_TK_NOTEXISTS
	}

	_, err = conn.Do("HMSET", "tk:"+hash, "key", key, "exp", expire+time.Now().Unix())
	if err != nil {
		return constants.ERR_INT_TK_DB
	} else {
		conn.Do("EXPIRE", "tk:"+hash, expire)
		return constants.ERR_INT_OK
	}
}

func (r *RedisDB) Delete(hash string) int {
	conn := common.GetRedisConn()
	defer conn.Close()

	_, err := conn.Do("DEL", "tk:"+hash)
	if err != nil {
		return constants.ERR_INT_TK_DB
	} else {
		return constants.ERR_INT_OK
	}
}

func (r *RedisDB) getField(hash string, field string) (string, int) {
	conn := common.GetRedisConn()
	defer conn.Close()

	reply, err := conn.Do("HGET", "tk:"+hash, field)
	if err != nil {
		return "", constants.ERR_INT_TK_DB
	} else if reply == nil {
		return "", constants.ERR_INT_TK_NOTEXISTS
	} else {
		val, _ := redis.String(reply, nil)
		return val, constants.ERR_INT_OK
	}
}

func (r *RedisDB) GetUID(hash string) (string, int) {
	return r.getField(hash, "uid")
}

func (r *RedisDB) GetKey(hash string) (string, int) {
	return r.getField(hash, "key")
}

func (r *RedisDB) GetToken(hash string) (string, int) {
	return r.getField(hash, "token")
}

func (r *RedisDB) GetAll(hash string) (uid, key, token string, ret int) {
	conn := common.GetRedisConn()
	defer conn.Close()

	reply, err := redis.StringMap(conn.Do("HGETALL", "tk:"+hash))
	if err != nil {
		// logger.Info("------------------------, read cache error:", err)
		return "", "", "", constants.ERR_INT_TK_DB
	} else if len(reply) == 0 {
		// logger.Info("------------------------, read cache error: ERR_INT_TK_NOTEXISTS")
		return "", "", "", constants.ERR_INT_TK_NOTEXISTS
	} else {
		return reply["uid"], reply["key"], reply["token"], constants.ERR_INT_OK
	}

}

// func (r *RedisDB) GetTxID(key string) (int64, int) {
// 	conn := common.GetRedisConn()
// 	defer conn.Close()

// 	// idx, err := redis.Int(conn.Do("INCR", key))

// 	reply, err := conn.Do("INCR", key)
// 	if err != nil {
// 		return -1, constants.ERR_INT_TK_DB
// 	} else if reply == nil {
// 		return -1, constants.ERR_INT_TK_NOTEXISTS
// 	} else {
// 		idx, _ := redis.Int64(reply, nil)
// 		return idx, constants.ERR_INT_OK
// 	}
// }
