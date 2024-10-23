package repo

import (
	"context"
	"strconv"

	"codexie.com/w-book-article/internal/dao/cache"
	"codexie.com/w-book-article/internal/dao/db"
	"codexie.com/w-book-article/internal/domain"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/singleflight"
)

type IAuthorRepository interface {
	Save(ctx context.Context, article *domain.Article) (int64, error)
	SelectPage(ctx context.Context, authorId int64, page, size int) ([]*domain.Article, error)
	FindArticleById(ctx context.Context, id int64) (*domain.Article, error)
}

type AuthorRepository struct {
	articleDao   *db.AuthorDao
	articleCache cache.ArticleCache
	g            singleflight.Group
	isTx         bool
}

func NewAuthorRepository(articleDao *db.AuthorDao, cache cache.ArticleCache) IAuthorRepository {
	return &AuthorRepository{articleDao: articleDao, articleCache: cache}
}

func (artRepo *AuthorRepository) Save(ctx context.Context, article *domain.Article) (int64, error) {
	var (
		artEntity = ToEntity(article)
		err       error
	)
	if article.Id > 0 {
		err = artRepo.articleDao.UpdateById(ctx, artEntity)
	} else {
		err = artRepo.articleDao.Create(ctx, artEntity)
	}
	if err != nil {
		return 0, err
	}
	// 删除初始页缓存信息
	if err := artRepo.articleCache.DelFirstPage(ctx, artEntity.AuthorId); err != nil {
		logx.WithContext(ctx).Error("[Save] 删除首页缓存失败")
	}
	//异步缓存文章内容
	go func() {
		if err := artRepo.articleCache.CacheArticle(ctx, FromArticle(artEntity), false); err != nil {
			logx.Errorf("[Save] 异步缓存制作库文章失败,原因:%s", err)
		}
	}()
	return artEntity.Id, nil
}

func (artRepo *AuthorRepository) SelectPage(ctx context.Context, authorId int64, page, size int) ([]*domain.Article, error) {
	var res []*domain.Article
	if page == 1 {
		articles, err := artRepo.articleCache.GetFirstPage(context.Background(), authorId)
		if err != nil {
			logx.Errorf("[selectpage] 查询文章首页缓存失败,原因:%s", err)
		}
		if len(articles) != 0 {
			return articles, nil
		}
		key := "article:firstpage:" + strconv.Itoa(int(authorId))
		artRepo.g.Do(key, func() (interface{}, error) {
			// 缓存未命中，查询数据库
			articles, err := artRepo.articleDao.SelectPage(ctx, authorId, page, size)
			if err != nil {
				return nil, err
			}
			for _, article := range articles {
				res = append(res, FromArticle(article))
			}

			//缓存结果
			if len(res) != 0 {
				err = artRepo.articleCache.CacheFirstArtilePage(context.Background(), authorId, res)
			}
			if err != nil {
				logx.Errorf("[select page] 缓存文章首页失败,原因:%s", err)
			}
			return res, nil
		})
		return res, nil
	}
	articles, err := artRepo.articleDao.SelectPage(context.Background(), authorId, page, size)
	if err != nil {
		return nil, err
	}
	for _, article := range articles {
		res = append(res, FromArticle(article))
	}

	return res, nil
}

func (artRepo *AuthorRepository) FindArticleById(ctx context.Context, id int64) (*domain.Article, error) {
	res, err := artRepo.articleCache.GetArticleById(ctx, id, false)
	if err != nil {
		logx.Errorf("[AuthorRepository] 查询文章缓存失败,原因:%s", err)
	}
	if res != nil && res.Id != 0 {
		return res, nil
	}
	article, err := artRepo.articleDao.SelectById(ctx, id)
	if err != nil {
		return nil, err
	}
	return FromArticle(article), nil
}

func FromArticle(article *db.Article) *domain.Article {
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

func ToEntity(art *domain.Article) *db.Article {
	return &db.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
	}
}
