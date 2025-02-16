package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type ResourceCache interface {
	GetResourceIDs(ctx context.Context, biz string, offset, size int) ([]int64, error)
	SetResourceIDs(ctx context.Context, biz string, ids []int64, expireDuration time.Duration) error
}

type ResourceRedisCache struct {
	client *redis.Client
}

func NewResourceRedisCache(client *redis.Client) ResourceCache {
	return &ResourceRedisCache{client: client}
}

func (c *ResourceRedisCache) GetResourceIDs(ctx context.Context, biz string, offset, size int) ([]int64, error) {
	key := c.resourceKey(biz)
	result, err := c.client.LRange(ctx, key, int64(offset), int64(offset+size-1)).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "获取资源ID缓存失败, key=%s", key)
	}

	ids := make([]int64, 0, len(result))
	for _, v := range result {
		id, _ := strconv.ParseInt(v, 10, 64)
		ids = append(ids, id)
	}
	return ids, nil
}

func (c *ResourceRedisCache) SetResourceIDs(ctx context.Context, biz string, ids []int64, expireDuration time.Duration) error {
	key := c.resourceKey(biz)

	// 先删除旧缓存
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return errors.Wrapf(err, "删除旧资源ID缓存失败, key=%s", key)
	}

	// 转换类型并存储
	args := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		args = append(args, strconv.FormatInt(id, 10))
	}

	if err := c.client.RPush(ctx, key, args...).Err(); err != nil {
		return errors.Wrapf(err, "存储资源ID缓存失败, key=%s", key)
	}

	return c.client.Expire(ctx, key, expireDuration).Err()
}

func (c *ResourceRedisCache) resourceKey(biz string) string {
	return fmt.Sprintf("resource:%s:ids", biz)
}
