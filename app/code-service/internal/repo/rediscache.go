package repo

import (
	"codexie.com/w-book-user/pkg/common/codeerr"
	"context"
	"github.com/redis/go-redis/v9"
)

var (
	systemErr   = -2
	sendSuccess = 0
	sendBusy    = -1
)

type RedisCache struct {
	redisClient *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{redisClient: client}
}

func (c *RedisCache) StoreCode(ctx context.Context, key, val, script string) error {
	result, err := c.redisClient.Eval(ctx, script, []string{key}, val).Int()
	if err != nil {
		return err
	}
	if result != sendSuccess {
		if result == systemErr {
			return codeerr.WithCode(codeerr.CodeSystemERR, "%s exist but no expire time", key)
		} else {
			return codeerr.WithCode(codeerr.CodeFrequentErr, "%s send too frequently", key)
		}
	}
	return nil
}

func (c *RedisCache) VerifyCode(ctx context.Context, key, val, script string) error {
	result, err := c.redisClient.Eval(ctx, script, []string{key}, val).Int()
	if err != nil {
		return err
	}

	switch result {
	case -1:
		return codeerr.WithCode(codeerr.CodeNotExistErr, "%s not exist", key)
	case -2:
		return codeerr.WithCode(codeerr.CodeVerifyExcceddErr, "%s verify too much", key)
	case -3:
		return codeerr.WithCode(codeerr.CodeVerifyFailERR, "%s verify not match", key)
	default:
		return nil
	}
}
