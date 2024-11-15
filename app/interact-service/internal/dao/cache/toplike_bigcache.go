package cache

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"codexie.com/w-book-interact/internal/dao/db"
	"github.com/allegro/bigcache/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	topCnt = 100
)

/**
* localcache放在repo中，由repo来完成实际的更新
* 创建两个worker针对redis和本地缓存，使用for + select监听机制，监听两个通道分别是ticker和type变更通道。
**/
type TopLikeCacheBC struct {
	localCache    *bigcache.BigCache
	interactCache InteractCache
	locks         map[string]*sync.Mutex
	lockStatus    map[string]bool
	mu            sync.Mutex
}

func NewBigCacheResourceCache(cache InteractCache) (*TopLikeCacheBC, error) {
	bigCache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute)) // 10 minutes expiration
	if err != nil {
		return nil, err
	}

	return &TopLikeCacheBC{
		localCache:    bigCache,
		interactCache: cache,
		locks:         make(map[string]*sync.Mutex),
		lockStatus:    make(map[string]bool),
	}, nil
}

func (b *TopLikeCacheBC) UpdateFromRedis(resourceType string) error {
	// todo: 更新时使用本地锁锁住，若该资源已经上锁则直接返回
	var result []db.Interaction
	if !b.TryLock(resourceType) {
		logx.Infof("已有协程在更新资源[%s]的本地缓存,不再重复更新", resourceType)
		return nil
	}
	defer b.Unlock(resourceType)
	result = b.interactCache.GetTopLikeResources(resourceType, topCnt)

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return b.localCache.Set(resourceType, data)
}

func (b *TopLikeCacheBC) GetTopResources(resourceType string) ([]db.Interaction, error) {
	data, err := b.localCache.Get(resourceType)
	if err != nil {
		return nil, err
	}

	var resources []db.Interaction
	err = json.Unmarshal(data, &resources)
	if err != nil {
		return nil, err
	}

	// Sort resources by likes in descending order
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].LikeCnt > resources[j].LikeCnt
	})

	return resources, nil
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
