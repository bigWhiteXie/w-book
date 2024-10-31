package cache

import (
	"context"
	"errors"
	"strconv"
	"strings"
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
	cntMap, err := c.redisClient.HGetAll(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	arr := strings.Split(key, ":")
	bizId, _ := strconv.Atoi(arr[2])
	likeCnt, _ := strconv.Atoi(cntMap[domain.Like])
	readCnt, _ := strconv.Atoi(cntMap[domain.Read])
	colCnt, _ := strconv.Atoi(cntMap[domain.Collect])

	return &domain.StatCnt{
		Biz:        arr[1],
		BizId:      int64(bizId),
		LikeCnt:    int64(likeCnt),
		ReadCnt:    int64(readCnt),
		CollectCnt: int64(colCnt),
	}, nil
}

// 缓存资源计数信息
func (c *InteractRedis) CacheStatCnt(ctx context.Context, key string, info *domain.StatCnt) error {
	cntMap := map[string]int64{
		domain.Like:    info.LikeCnt,
		domain.Collect: info.CollectCnt,
		domain.Read:    info.ReadCnt,
	}
	if err := c.redisClient.HSet(ctx, key, cntMap).Err(); err != nil {
		return err
	}
	return c.redisClient.Expire(ctx, key, 30*time.Minute).Err()
}
