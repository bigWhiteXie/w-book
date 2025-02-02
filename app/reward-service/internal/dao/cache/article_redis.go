package cache

import (
	"github.com/redis/go-redis/v9"
)

var (
	firstPageKey     = "article:firstpage:"
	articleAuthorKey = "article:author:"
	articleReaderKey = "article:reader:"
)

type RedisCache struct {
	redisClient *redis.Client
}

func NewRedisCache(client *redis.Client) PaymentCache {
	return &RedisCache{redisClient: client}
}
