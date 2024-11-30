package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"codexie.com/w-book-article/internal/domain"
	"github.com/pkg/errors"
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
	if err := c.redisClient.Set(ctx, firstPageKey+strconv.Itoa(int(authorId)), domain.ArticleArray(list), 6*time.Hour).Err(); err != nil {
		return errors.Wrapf(err, "[ArticleRedis_CacheFirstArtilePage] 缓存作者库首页文章列表失败")
	}

	return nil
}

func (c *ArticleRedis) CacheArticle(ctx context.Context, article *domain.Article, isPublish bool) error {
	key := articleAuthorKey + strconv.Itoa(int(article.Id))
	if isPublish {
		key = articleReaderKey + strconv.Itoa(int(article.Id))
	}
	if err := c.redisClient.Set(ctx, key, article, 10*time.Minute).Err(); err != nil {
		return errors.Wrapf(err, "[ArticleRedis_CacheArticle] 缓存文章失败,article_id=%d", article.Id)
	}

	return nil
}

func (c *ArticleRedis) GetArticleById(ctx context.Context, articleId int64, isPublish bool) (*domain.Article, error) {
	res := &domain.Article{}
	key := articleAuthorKey + strconv.Itoa(int(articleId))
	if isPublish {
		key = articleReaderKey + strconv.Itoa(int(articleId))
	}
	bytes, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, errors.Wrapf(err, "[ArticleRedis_GetArticleById] 获取文章失败,article_id=%d", articleId)
	}
	if err := json.Unmarshal(bytes, res); err != nil {
		return nil, errors.Wrapf(err, "[ArticleRedis_GetArticleById] 反序列化文章失败,article_id=%d", articleId)
	}

	return res, nil
}

func (c *ArticleRedis) GetFirstPage(ctx context.Context, authorId int64) ([]*domain.Article, error) {
	bytes, err := c.redisClient.Get(ctx, firstPageKey+strconv.Itoa(int(authorId))).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, errors.Wrapf(err, "[ArticleRedis_GetFirstPage] 获取文章失败,author_id=%d", authorId)
	}
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	var list domain.ArticleArray
	if err = json.Unmarshal(bytes, &list); err != nil {
		return nil, errors.Wrapf(err, "[ArticleRedis_GetFirstPage] 反序列化文章失败,author_id=%d", authorId)
	}

	return []*domain.Article(list), nil
}

func (c *ArticleRedis) DelFirstPage(ctx context.Context, authorId int64) error {
	if err := c.redisClient.Del(ctx, firstPageKey+strconv.Itoa(int(authorId))).Err(); err != nil {
		return errors.Wrapf(err, "[ArticleRedis_DelFirstPage] 删除文章失败,author_id=%d", authorId)
	}

	return nil
}
