package common

import (
	"time"

	"utils/config"
	"utils/logger"

	"github.com/garyburd/redigo/redis"
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
			logger.Error("redis conn err ",err.Error())
		}
		return conn
	}
	return nil
}
