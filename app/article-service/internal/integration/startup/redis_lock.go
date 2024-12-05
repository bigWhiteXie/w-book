package startup

import (
	red "github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
)

func InitRedLock() *redsync.Redsync {
	rdb := red.NewClient(&red.Options{
		Addr:     "127.0.0.1:16379",
		Password: "",
		DB:       0,
	})
	pool := goredis.NewPool(rdb)

	return redsync.New(pool)
}
