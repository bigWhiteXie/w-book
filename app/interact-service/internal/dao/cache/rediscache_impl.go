package cache

import (
	"context"
	"strconv"
	"strings"
	"time"

	"codexie.com/w-book-interact/internal/domain"
	"github.com/go-redsync/redsync/v4"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	firstPageKey     = "article:firstpage:"
	articleAuthorKey = "article:author:"
	articleReaderKey = "article:reader:"

	resourceLockKey = "resource:liketop:lock:"
)

type InteractRedis struct {
	redisClient *redis.Client
	redLock     *redsync.Redsync
}

func NewInteractRedis(client *redis.Client, rs *redsync.Redsync) InteractCache {
	return &InteractRedis{redisClient: client, redLock: rs}
}

func (c *InteractRedis) IncreCntIfExist(ctx context.Context, key string, cntKind string, increment int) error {
	result, err := c.redisClient.Eval(ctx, updateCntTemplate(), []string{key}, cntKind, increment).Int()
	if result == 1 {
		logx.Infof("成功更新交互缓存计数,key:%s,类型:%s", key, cntKind)
	} else {
		logx.Infof("更新交互缓存计数失败,key:%s不存在", key)
	}

	return errors.Wrapf(err, "[InteractRedis_IncreCntIfExist] 更新交互缓存计数失败, key=%s", key)
}

// 查询资源计数缓存
func (c *InteractRedis) GetStatCnt(ctx context.Context, key string) (*domain.Interaction, error) {
	cntMap, err := c.redisClient.HGetAll(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "[InteractRedis_GetStatCnt] 获取资源计数缓存计数失败, key=%s", key)
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

func (c *InteractRedis) UpdateRedisZSet(ctx context.Context, resourceType string, fn func() ([]*domain.Interaction, error)) error {
	zsetKey := "resource:top:" + resourceType
	lockKey := resourceLockKey + resourceType
	mutex := c.redLock.NewMutex(lockKey,
		redsync.WithExpiry(10*time.Second), // 锁过期时间
		redsync.WithTries(1),
	)

	if err := mutex.TryLock(); err != nil {
		if err == redsync.ErrFailed {
			logx.Infof("当前其它服务正在占用锁:%s", lockKey)
			return nil
		}
		logx.Errorf("获取分布式锁%s失败", lockKey)
		return errors.Wrapf(err, "[InteractRedis_UpdateRedisZSet] 获取分布式锁失败,lockKey=%s", lockKey)
	}
	// 获取分布式锁成功，更新zset
	defer mutex.Unlock()
	resources, err := fn()
	logx.Infof("从数据库中拿到资源%s,长度%d", resourceType, len(resources))
	if err != nil {
		return err
	}
	args := []interface{}{}
	for _, resource := range resources {
		args = append(args, strconv.FormatInt(resource.BizId, 10))
		args = append(args, strconv.FormatInt(resource.LikeCnt, 10))
	}
	if err := c.redisClient.Eval(ctx, updateTopLike(), []string{zsetKey}, args...).Err(); err != nil {
		return errors.Wrapf(err, "[InteractRedis_UpdateRedisZSet] 更新redis zset失败,key=%s", zsetKey)
	}
	return nil
}

func (c *InteractRedis) IncrementLikeInZSet(ctx context.Context, resourceType string, resourceId int64, incre int) error {
	zsetKey := "resource:top:" + resourceType
	if err := c.redisClient.Eval(ctx, increResourceLike(), []string{zsetKey}, strconv.FormatInt(resourceId, 10), incre).Err(); err != nil {
		return errors.Wrapf(err, "[InteractRedis_IncrementLikeInZSet] 更新redis zset失败,key=%s", zsetKey)
	}

	return nil
}

func (c *InteractRedis) GetTopFromRedisZSet(ctx context.Context, resourceType string, limit int) ([]int64, error) {
	zsetKey := "resource:top:" + resourceType
	strIds, err := c.redisClient.ZRevRangeByScore(ctx, zsetKey, &redis.ZRangeBy{Min: "0", Max: "+inf", Count: int64(limit)}).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "[InteractRedis_GetTopFromRedisZSet] 获取redis zset失败,key=%s", zsetKey)
	}
	res := make([]int64, 0, len(strIds))
	for _, strId := range strIds {
		id, _ := strconv.Atoi(strId)
		res = append(res, int64(id))
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
		return errors.Wrapf(err, "[InteractRedis_CacheStatCnt] 缓存资源计数信息失败,key=%s", key)
	}
	if err := c.redisClient.Expire(ctx, key, 30*time.Minute).Err(); err != nil {
		return errors.Wrapf(err, "[InteractRedis_CacheStatCnt] 设置资源计数信息过期时间失败,key=%s", key)
	}

	return nil
}
