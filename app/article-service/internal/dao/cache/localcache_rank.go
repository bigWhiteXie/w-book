package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"codexie.com/w-book-article/internal/domain"
	"github.com/allegro/bigcache/v3"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

var topKey = "rank:top:" + domain.Biz

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

func (cache *LocalArtTopCache) ReplaceTopN(ctx context.Context, arts []*domain.Article) error {
	// 将文章列表转换为适合传递给Lua脚本的参数格式，这里进行json序列化
	value, err := json.Marshal(arts)
	if err != nil {
		return errors.Wrapf(err, "[LocalArtTopCache_ReplaceTopN] 序列化文章结构体失败：%s", err)
	}
	cache.rw.Lock()
	defer cache.rw.Unlock()
	if err := cache.localCache.Set(topKey, value); err != nil {
		logx.Errorf("放入本地缓存失败:%s", err)
		return errors.Wrapf(err, "[LocalArtTopCache_ReplaceTopN] 放入本地缓存失败:%s", err)
	}

	return nil
}

func (c *LocalArtTopCache) TakeTopNArticles(ctx context.Context) ([]*domain.Article, error) {
	var articles domain.ArticleArray
	c.rw.RLock()
	defer c.rw.RUnlock()
	byteData, err := c.localCache.Get(topKey)
	if err != nil {
		return nil, errors.Wrapf(err, "[LocalArtTopCache_TakeTopNArticles] 从缓存中取出数据失败:%s", err)
	}

	json.Unmarshal(byteData, &articles)
	return articles, nil
}
