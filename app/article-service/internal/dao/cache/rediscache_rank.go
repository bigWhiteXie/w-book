package cache

import (
	"context"
	"encoding/json"

	"codexie.com/w-book-article/internal/domain"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type RedisArtTopNCache struct {
	redisClient *redis.Client
}

func NewRankCacheRedis(client *redis.Client) *RedisArtTopNCache {
	return &RedisArtTopNCache{redisClient: client}
}

func (c *RedisArtTopNCache) ReplaceTopN(ctx context.Context, arts []*domain.Article) error {
	topKey := "rank:top:" + domain.Biz
	luaScript := `
    -- 删除指定key对应的List
    redis.call('DEL', KEYS[1])
    -- 遍历要添加的文章列表, 逐个添加到List中
    for i, art in ipairs(ARGV) do
        local serialized_art = art
        redis.call('RPUSH', KEYS[1], serialized_art)
    end
    return 1
    `
	// 将文章列表转换为适合传递给Lua脚本的参数格式，这里进行json序列化
	args := make([]interface{}, len(arts))
	for i, art := range arts {
		serializedArt, err := json.Marshal(art)
		if err != nil {
			return errors.Wrapf(err, "[Redis_ReplaceTopN] 序列化文章结构体失败，文章信息：%+v", art)
		}
		args[i] = string(serializedArt)
	}
	if err := c.redisClient.Eval(ctx, luaScript, []string{topKey}, args...).Err(); err != nil {
		return errors.Wrapf(err, "[Redis_ReplaceTopN] 更新redis LIST失败,key=%s,err:%s", topKey, err)
	}

	return nil
}

func (c *RedisArtTopNCache) TakeTopNArticles(ctx context.Context) ([]*domain.Article, error) {
	topKey := "rank:top:" + domain.Biz
	// 从Redis List中获取所有元素
	result, err := c.redisClient.LRange(ctx, topKey, 0, -1).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "[Redis_TakeTopNArticles] 从redis LIST获取数据失败,key=%s", topKey)
	}
	articles := make([]*domain.Article, len(result))
	for i, serializedArt := range result {
		var art domain.Article
		err := json.Unmarshal([]byte(serializedArt), &art)
		if err != nil {
			return nil, errors.Wrapf(err, "[Redis_TakeTopNArticles] 反序列化文章结构体失败，序列化数据：%s", serializedArt)
		}
		articles[i] = &art
	}
	return articles, nil
}
