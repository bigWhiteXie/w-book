package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"codexie.com/w-book-interact/internal/domain"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	firstPageKey     = "article:firstpage:"
	articleAuthorKey = "article:author:"
	articleReaderKey = "article:reader:"
)

type InteractRedis struct {
	redisClient *redis.Client
}

func NewInteractRedis(client *redis.Client) InteractCache {
	return &InteractRedis{redisClient: client}
}

func (c *InteractRedis) IncreCntIfExist(ctx context.Context, key string, cntKind string, increment int) error {
	result, err := c.redisClient.Eval(ctx, updateCntTemplate(), []string{key}, cntKind, increment).Int()
	if result == 1 {
		logx.Infof("成功更新点赞缓存计数,key:%s", key)
	} else {
		logx.Infof("更新点赞缓存计数失败,key:%s不存在", key)
	}
	return err
}

// 查询资源计数缓存
func (c *InteractRedis) GetStatCnt(ctx context.Context, key string) (*domain.StatCnt, error) {
	bytes, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	var statCnt *domain.StatCnt
	err = json.Unmarshal(bytes, statCnt)

	return statCnt, nil
}

// 缓存资源计数信息
func (c *InteractRedis) CacheStatCnt(ctx context.Context, key string, info *domain.StatCnt) error {
	return c.redisClient.Set(ctx, key, info, 30*time.Minute).Err()
}
