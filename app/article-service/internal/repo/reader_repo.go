package repo

import (
	"context"
	"time"

	"codexie.com/w-book-article/internal/dao"
	"codexie.com/w-book-article/internal/domain"
)

type IReaderRepository interface {
	Save(ctx context.Context, article *domain.Article) (int64, error)
	FindById(ctx context.Context, id int64) (*domain.Article, error)
}

type ReaderRepository struct {
	readerDao *dao.ReaderDao
	isTx      bool
}

func NewReaderRepository(readerDao *dao.ReaderDao) IReaderRepository {
	return &ReaderRepository{readerDao: readerDao}
}

func (artRepo *ReaderRepository) Save(ctx context.Context, article *domain.Article) (int64, error) {
	now := time.Now().UnixMilli()
	publishedArtcile := &dao.PublishedArticle{
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
	return publishedArtcile.Id, nil
}

func (artRepo *ReaderRepository) FindById(ctx context.Context, id int64) (*domain.Article, error) {
	article, err := artRepo.readerDao.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	return FromPublishedArticle(article), nil
}

func FromPublishedArticle(article *dao.PublishedArticle) *domain.Article {
	return &domain.Article{
		Id:      article.Id,
		Title:   article.Title,
		Content: article.Content,
		Status:  domain.ArticleStatusFromUint8(article.Status),
		Author: domain.Author{
			Id: article.AuthorId,
		},
	}
}
