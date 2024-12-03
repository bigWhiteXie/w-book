package cache

import (
	"context"

	"codexie.com/w-book-common/codeerr"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	systemErr   = -2
	sendSuccess = 0
	sendBusy    = -1
)

type CodeRedisCache struct {
	redisClient *redis.Client
	db          *gorm.DB
}

func NewCodeRedisCache(client *redis.Client) *CodeRedisCache {
	return &CodeRedisCache{redisClient: client}
}

func (c *CodeRedisCache) StoreCode(ctx context.Context, key, val, script string) error {
	result, err := c.redisClient.Eval(ctx, script, []string{key}, val).Int()
	if err != nil {
		return err
	}
	if result != sendSuccess {
		if result == systemErr {
			return errors.Wrap(codeerr.WithCode(codeerr.CodeSystemERR, "[CodeRedisCache_StoreCode]验证码key=%s存在但没过期时间", key), "")
		} else {
			return errors.Wrap(codeerr.WithCode(codeerr.CodeFrequentErr, "[CodeRedisCache_StoreCode]key=%s,验证码发送太频繁", key), "")
		}
	}
	return nil
}

func (c *CodeRedisCache) VerifyCode(ctx context.Context, key, val, script string) error {
	result, err := c.redisClient.Eval(ctx, script, []string{key}, val).Int()
	if err != nil {
		return err
	}

	switch result {
	case -1:
		return errors.Wrap(codeerr.WithCode(codeerr.CodeNotExistErr, "[CodeRedisCache_VerifyCode]验证码key=%s不存在", key), "")
	case -2:
		return errors.Wrap(codeerr.WithCode(codeerr.CodeVerifyExcceddErr, "[CodeRedisCache_VerifyCode]验证码key=%s校验次数过多", key), "")
	case -3:
		return errors.Wrap(codeerr.WithCode(codeerr.CodeVerifyFailERR, "[CodeRedisCache_VerifyCode]验证码key=%s校验不匹配", key), "")
	default:
		return nil
	}
}
