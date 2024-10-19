package cache

import (
	"context"

	"codexie.com/w-book-article/internal/domain"
)

type ArticleCache interface {
	//缓存制作库第一页文章信息
	CacheFirstArtilePage(ctx context.Context, authorId int64, list []*domain.Article) error

	//缓存文章
	CacheArticle(ctx context.Context, article *domain.Article, isPublish bool) error

	//读取文章内容
	GetArticleById(ctx context.Context, articleId int64, isPublish bool) (*domain.Article, error)

	//读取第一页文章信息列表
	GetFirstPage(ctx context.Context, authorId int64) ([]*domain.Article, error)

	//删除第一页信息
	DelFirstPage(ctx context.Context, authorId int64) error
}
