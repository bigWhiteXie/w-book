package startup

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	myRedis := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:16379",
		Password: "",
		DB:       0,
	})

	_, err := myRedis.Ping(context.Background()).Result()
	if err != nil {
		panic("redis connect failed: " + err.Error())
	}
	return myRedis
}
