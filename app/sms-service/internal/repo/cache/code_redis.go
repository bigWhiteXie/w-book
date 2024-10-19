package cache

import (
	"context"

	"codexie.com/w-book-common/common/codeerr"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	systemErr   = -2
	sendSuccess = 0
	sendBusy    = -1
)

type RedisCache struct {
	redisClient *redis.Client
	db          *gorm.DB
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
			return codeerr.WithCode(codeerr.CodeSystemERR, "验证码key存在但没过期时间", key)
		} else {
			return codeerr.WithCode(codeerr.CodeFrequentErr, "验证码发送太频繁", key)
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
		return codeerr.WithCode(codeerr.CodeNotExistErr, "验证码不存在")
	case -2:
		return codeerr.WithCode(codeerr.CodeVerifyExcceddErr, "验证码校验次数过多")
	case -3:
		return codeerr.WithCode(codeerr.CodeVerifyFailERR, "验证码不匹配")
	default:
		return nil
	}
}
