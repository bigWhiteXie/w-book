package repo

import (
	"context"
	"strconv"
	"time"

	"codexie.com/w-book-article/internal/dao/cache"
	"codexie.com/w-book-article/internal/dao/db"
	"codexie.com/w-book-article/internal/domain"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/singleflight"
)

type IReaderRepository interface {
	Save(ctx context.Context, article *domain.Article) (int64, error)
	FindById(ctx context.Context, id int64) (*domain.Article, error)
}

type ReaderRepository struct {
	readerDao    *db.ReaderDao
	articleCache cache.ArticleCache
	sg           singleflight.Group
	isTx         bool
}

func NewReaderRepository(readerDao *db.ReaderDao, cache cache.ArticleCache) IReaderRepository {
	return &ReaderRepository{readerDao: readerDao, articleCache: cache}
}

func (artRepo *ReaderRepository) Save(ctx context.Context, article *domain.Article) (int64, error) {
	now := time.Now().UnixMilli()
	publishedArtcile := &db.PublishedArticle{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   uint8(article.Status),
		Ctime:    now,
		Utime:    now,
	}
	if err := artRepo.readerDao.Save(ctx, publishedArtcile); err != nil {
		return 0, err
	}

	if err := artRepo.articleCache.CacheArticle(ctx, FromPublishedArticle(publishedArtcile), true); err != nil {
		logx.WithContext(ctx).Errorf("[Save] 缓存线上库文章失败,原因:%s", err)
	}
	return publishedArtcile.Id, nil
}

func (artRepo *ReaderRepository) FindById(ctx context.Context, id int64) (*domain.Article, error) {
	res, err := artRepo.articleCache.GetArticleById(ctx, id, true)
	if err != nil {
		logx.Errorf("[FindArticleById] 查询缓存失败,原因:%s", err)
	}
	if err == nil && res != nil {
		return res, nil
	}

	result, err, _ := artRepo.sg.Do(
		strconv.Itoa(int(id)),
		func() (interface{}, error) {
			// 缓存未命中，查询数据库
			article, err := artRepo.readerDao.FindById(ctx, id)
			if err != nil {
				return nil, err
			}

			// 将结果放入缓存
			if err := artRepo.articleCache.CacheArticle(ctx, FromPublishedArticle(article), true); err != nil {
				logx.Errorf("[FindArticleById] 缓存文章失败,原因:%s", err)
			}
			// 查询成功，将结果存入缓存（可选）
			// 你可以根据需要调用 articleCache.CacheFirstPage 或其他缓存方法
			return FromPublishedArticle(article), nil
		},
	)

	if err != nil {
		return nil, err
	}

	return result.(*domain.Article), nil
}

func FromPublishedArticle(article *db.PublishedArticle) *domain.Article {
	return &domain.Article{
		Id:      article.Id,
		Title:   article.Title,
		Content: article.Content,
		Status:  domain.ArticleStatusFromUint8(article.Status),
		Author: domain.Author{
			Id: article.AuthorId,
		},
		Ctime: article.Ctime,
		Utime: article.Utime,
	}
}
