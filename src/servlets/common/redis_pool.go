package common

import (
	"time"

	"utils/logger"

	"github.com/garyburd/redigo/redis"
)

var redisPool *redis.Pool

func Init_redis(conf map[string]string) {
	logger.Debug(conf)
	redisPool = &redis.Pool{
		MaxIdle:     8,
		MaxActive:   16,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			//c, err := redis.Dial("tcp", conf["addr"])
			c, err := redis.Dial("tcp", conf["addr"],
				redis.DialConnectTimeout(500*time.Millisecond),
				redis.DialReadTimeout(500*time.Millisecond),
				redis.DialWriteTimeout(500*time.Millisecond),
				redis.DialKeepAlive(1*time.Second),
				redis.DialPassword(conf["auth"]))
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
		return redisPool.Get()
	}

	return nil
}
