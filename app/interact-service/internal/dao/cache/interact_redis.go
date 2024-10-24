package cache

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"codexie.com/w-book-interact/internal/dao/db"
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
func (c *InteractRedis) GetStatCnt(ctx context.Context, key string) (*domain.Interaction, error) {
	cntMap, err := c.redisClient.HGetAll(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(cntMap) == 0 {
		return nil, nil
	}
	arr := strings.Split(key, ":")
	bizId, _ := strconv.Atoi(arr[2])
	likeCnt, _ := strconv.Atoi(cntMap[domain.Like])
	readCnt, _ := strconv.Atoi(cntMap[domain.Read])
	colCnt, _ := strconv.Atoi(cntMap[domain.Collect])

	return &domain.Interaction{
		Biz:        arr[1],
		BizId:      int64(bizId),
		LikeCnt:    int64(likeCnt),
		ReadCnt:    int64(readCnt),
		CollectCnt: int64(colCnt),
	}, nil
}

func (c *InteractRedis) UpdateRedisZSet(ctx context.Context, resourceType string, resources []db.Interaction) error {
	zsetKey := "resource:top:" + resourceType
	zset := make([]redis.Z, len(resources))
	for i, res := range resources {
		zset[i] = redis.Z{Score: float64(res.LikeCnt), Member: res}
	}
	return c.redisClient.ZAdd(ctx, zsetKey, zset...).Err()
}

func (c *InteractRedis) IncrementLikeInZSet(ctx context.Context, resourceType string, resourceID int64, incre int) error {
	zsetKey := "resource:top:" + resourceType
	return c.redisClient.ZIncrBy(ctx, zsetKey, float64(incre), string(resourceID)).Err()
}

func (c *InteractRedis) GetTopFromRedisZSet(ctx context.Context, resourceType string, limit int) ([]db.Interaction, error) {
	zsetKey := "resource:top:" + resourceType
	resources, err := c.redisClient.ZRevRangeByScore(ctx, zsetKey, &redis.ZRangeBy{Min: "0", Max: "+inf", Count: int64(limit)}).Result()
	if err != nil {
		return nil, err
	}
	res := make([]db.Interaction, len(resources))
	for _, resource := range resources {
		entity := &db.Interaction{}
		if err := json.Unmarshal([]byte(resource), entity); err == nil {
			res = append(res, *entity)
		} else {
			logx.Error("资源缓存反序列化失败:%s", err)
		}
	}

	return res, nil
}

// 缓存资源计数信息
func (c *InteractRedis) CacheStatCnt(ctx context.Context, key string, info *domain.Interaction) error {
	cntMap := map[string]string{
		domain.Like:    strconv.Itoa(int(info.LikeCnt)),
		domain.Collect: strconv.Itoa(int(info.CollectCnt)),
		domain.Read:    strconv.Itoa(int(info.ReadCnt)),
	}
	if err := c.redisClient.HSet(ctx, key, cntMap).Err(); err != nil {
		return err
	}
	return c.redisClient.Expire(ctx, key, 30*time.Minute).Err()
}
