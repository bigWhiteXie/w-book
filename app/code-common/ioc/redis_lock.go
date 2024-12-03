package ioc

import (
	red "github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
)

func InitRedLock(redisConf RedisConf) *redsync.Redsync {
	rdb := red.NewClient(&red.Options{
		Addr:     redisConf.Host,
		Password: redisConf.Pass,
		DB:       redisConf.DB,
	})
	pool := goredis.NewPool(rdb)

	return redsync.New(pool)
}
