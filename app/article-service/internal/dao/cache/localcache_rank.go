package cache

import (
	"context"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
)

type LocalArtTopCache struct {
	localCache *bigcache.BigCache
	rw         sync.RWMutex
}

func NewLocalArtTopCache() *LocalArtTopCache {
	bigCache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute)) // 10 minutes expiration
	if err != nil {
		panic("[LocalArtTopCache] 初始化本地缓存失败:" + err.Error())
	}
	return &LocalArtTopCache{
		localCache: bigCache,
	}
}
