package common

import (
	"time"

	"utils/config"
	"utils/logger"

	"errors"
	"github.com/garyburd/redigo/redis"
	"utils"
)

var redisPool *redis.Pool

func RedisPoolInit() {
	redisCfg := config.GetConfig().Redis
	redisPool = &redis.Pool{
		MaxIdle:     redisCfg.MaxConn,
		MaxActive:   redisCfg.MaxConn,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			//c, err := redis.Dial("tcp", conf["addr"])
			c, err := redis.Dial("tcp", redisCfg.RedisAddr,
				redis.DialConnectTimeout(500*time.Millisecond),
				redis.DialReadTimeout(500*time.Millisecond),
				redis.DialWriteTimeout(500*time.Millisecond),
				redis.DialKeepAlive(20*time.Second),
				redis.DialPassword(redisCfg.RedisAuth))
			if err != nil {
				logger.Info("token: can't connect to redis server")
				return nil, err
			}

			// if len(conf["auth"]) > 0 {
			// 	succ, err := redis.Bool(c.Do("AUTH", conf["auth"]))
			// 	if err != nil {
			// 		logger.Info("token: can't connect to redis server")
			// 		c.Close()
			// 		return nil, err
			// 	} else if !succ {
			// 		logger.Info("token: redis server password wrong")
			// 		c.Close()
			// 		return nil, errors.New("redis server password wrong")
			// 	}
			// }

			if redisCfg.DBIndex > 0 {
				succ, err := redis.String(c.Do("SELECT", redisCfg.DBIndex))
				if err != nil {
					logger.Info("select db",redisCfg.DBIndex,"failed",err.Error())
					c.Close()
					return nil, err
				} else if succ != "OK" {
					logger.Info("select db",redisCfg.DBIndex,"failed")
					c.Close()
					return nil, errors.New("can not select db " + utils.Int2Str(redisCfg.DBIndex))
				}
				//logger.Info("select db res",succ)
			}

			//logger.Info("select db",redisCfg.DBIndex)
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func GetRedisConn() redis.Conn {
	if redisPool != nil {
		conn := redisPool.Get()
		err := conn.Err()
		if err != nil {
			logger.Error("redis conn err ", err.Error())
		}
		return conn
	}
	return nil
}

func rdsDo(commandName string, args ...interface{}) (reply interface{}, err error) {
	conn := GetRedisConn()
	if conn == nil {
		return 0, errors.New("can not connect redis")
	}
	defer conn.Close()
	return conn.Do(commandName, args...)
}

func ttl(key string) (int, error) {
	return redis.Int(rdsDo("TTL", key))
}

func incr(key string) (int, error) {
	return redis.Int(rdsDo("INCR", key))
}

func incrby(key string, value int64) (int, error) {
	return redis.Int(rdsDo("INCRBY", key, value))
}

func rdsGet(key string) (int, error) {
	return redis.Int(rdsDo("GET", key))
}

func rdsGet64(key string) (int64, error) {
	return redis.Int64(rdsDo("GET", key))
}

func rdsDel(key string) error {
	_, err := rdsDo("DEL", key)
	return err
}

func rdsExpire(key string, expire int) error {
	_, err := rdsDo("EXPIRE", key, expire)
	return err
}

func setAndExpire(key string, value, expire int) error {
	_, err := rdsDo("SET", key, value, "EX", expire)
	return err
}

func setAndExpire64(key string, value int64, expire int) error {
	_, err := rdsDo("SET", key, value, "EX", expire)
	return err
}

func setnx(key string,value int64) (int,error) {
	return redis.Int(rdsDo("SETNX",key,value))
}

func hmset(key string, p map[string]string) (string, error) {
	args := []interface{}{key}
	for i, v := range p {
		args = append(args, i, v)
	}
	return redis.String(rdsDo("HMSET", args...))
}

func hset(key string, fieldName,fieldValue string) (string, error) {
	return redis.String(rdsDo("HSET", fieldName,fieldValue))
}

func hgetall(key string) (map[string]string, error) {
	return redis.StringMap(rdsDo("HGETALL", key))
}

func hget(key string, fieldName string) (string, error) {
	return redis.String(rdsDo("HGET", fieldName))
}
