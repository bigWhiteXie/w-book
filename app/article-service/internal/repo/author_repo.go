package repo

import (
	"context"

	"codexie.com/w-book-article/internal/dao"
	"codexie.com/w-book-article/internal/domain"
)

type IAuthorRepository interface {
	Save(ctx context.Context, article *domain.Article) (int64, error)
}

type AuthorRepository struct {
	articleDao *dao.AuthorDao
	isTx       bool
}

func NewAuthorRepository(articleDao *dao.AuthorDao) IAuthorRepository {
	return &AuthorRepository{articleDao: articleDao}
}

func (artRepo *AuthorRepository) Save(ctx context.Context, article *domain.Article) (int64, error) {
	var (
		artEntity = article.ToEntity()
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
	return artEntity.Id, nil
}
