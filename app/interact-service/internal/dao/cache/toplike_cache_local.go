package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

/**
* localcache放在repo中，由repo来完成实际的更新
* 创建两个worker针对redis和本地缓存，使用for + select监听机制，监听两个通道分别是ticker和type变更通道。
**/
type TopLikeCacheBC struct {
	localCache *bigcache.BigCache
	locks      map[string]*sync.Mutex
	lockStatus map[string]bool
	mu         sync.Mutex
}

func NewBigCacheResourceCache() TopLikeCache {
	bigCache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute)) // 10 minutes expiration
	if err != nil {
		panic(err)
	}

	return &TopLikeCacheBC{
		localCache: bigCache,
		locks:      make(map[string]*sync.Mutex),
		lockStatus: make(map[string]bool),
	}
}

/**
* 从redis中获取指定资源类型点赞数最高的前limit个资源，内部使用互斥锁，获取锁失败则直接返回
**/
func (b *TopLikeCacheBC) UpdateResourceCache(resourceType string, resourceIds []int64) error {
	data, err := json.Marshal(resourceIds)
	if err != nil {
		return errors.Wrapf(err, "[TopLikeCacheBC] 序列化失败,resourceType=%s, resourceIds=%v", resourceType, resourceIds)
	}
	logx.Infof("更新资源%s本地缓存,长度%d", resourceType, len(resourceIds))
	if err := b.localCache.Set(resourceType, data); err != nil {
		return errors.Wrapf(err, "[TopLikeCacheBC] 更新TopLike本地缓存失败,resourceType=%s, resourceIds=%v", resourceType, resourceIds)
	}
	return nil
}

func (b *TopLikeCacheBC) GetTopResources(resourceType string) ([]int64, error) {
	var resourceIds []int64

	data, err := b.localCache.Get(resourceType)
	if err != nil {
		return nil, errors.Wrapf(err, "[TopLikeCacheBC] 获取TopLike本地缓存失败,resourceType=%s", resourceType)
	}

	err = json.Unmarshal(data, &resourceIds)
	if err != nil {
		return nil, errors.Wrapf(err, "[TopLikeCacheBC] 反序列化TopLike本地缓存失败,resourceType=%s", resourceType)
	}

	return resourceIds, nil
}

func (b *TopLikeCacheBC) TryLock(resourceType string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.locks[resourceType]; !exists {
		b.locks[resourceType] = &sync.Mutex{}
		b.lockStatus[resourceType] = false
	}

	if b.lockStatus[resourceType] {
		return false
	}

	b.locks[resourceType].Lock()
	b.lockStatus[resourceType] = true
	return true
}

func (b *TopLikeCacheBC) Unlock(resourceType string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if lock, exists := b.locks[resourceType]; exists && b.lockStatus[resourceType] {
		lock.Unlock()
		b.lockStatus[resourceType] = false
	}
}
