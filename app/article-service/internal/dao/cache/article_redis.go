package cache

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"codexie.com/w-book-article/internal/domain"
	"github.com/redis/go-redis/v9"
)

var (
	firstPageKey     = "article:firstpage:"
	articleAuthorKey = "article:author:"
	articleReaderKey = "article:reader:"
)

type ArticleRedis struct {
	redisClient *redis.Client
}

func NewArticleRedis(client *redis.Client) ArticleCache {
	return &ArticleRedis{redisClient: client}
}

func (c *ArticleRedis) CacheFirstArtilePage(ctx context.Context, authorId int64, list []*domain.Article) error {
	return c.redisClient.Set(ctx, firstPageKey+strconv.Itoa(int(authorId)), domain.ArticleArray(list), 6*time.Hour).Err()
}

func (c *ArticleRedis) CacheArticle(ctx context.Context, article *domain.Article, isPublish bool) error {
	key := articleAuthorKey + strconv.Itoa(int(article.Id))
	if isPublish {
		key = articleReaderKey + strconv.Itoa(int(article.Id))
	}
	return c.redisClient.Set(ctx, key, article, 10*time.Minute).Err()
}

func (c *ArticleRedis) GetArticleById(ctx context.Context, articleId int64, isPublish bool) (*domain.Article, error) {
	res := &domain.Article{}
	key := articleAuthorKey + strconv.Itoa(int(articleId))
	if isPublish {
		key = articleReaderKey + strconv.Itoa(int(articleId))
	}
	bytes, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, res)
	return res, err
}

func (c *ArticleRedis) GetFirstPage(ctx context.Context, authorId int64) ([]*domain.Article, error) {
	bytes, err := c.redisClient.Get(ctx, firstPageKey+strconv.Itoa(int(authorId))).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	var list domain.ArticleArray
	err = json.Unmarshal(bytes, &list)
	return []*domain.Article(list), err
}

func (c *ArticleRedis) DelFirstPage(ctx context.Context, authorId int64) error {
	return c.redisClient.Del(ctx, firstPageKey+strconv.Itoa(int(authorId))).Err()
}
