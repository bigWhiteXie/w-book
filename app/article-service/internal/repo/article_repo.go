package repo

import "codexie.com/w-book-article/internal/dao"

type IArticleRepository interface {
}

type ArticleRepository struct {
	articleDao *dao.ArticleDao
	isTx       bool
}

func NewArticleRepository(articleDao *dao.ArticleDao) IArticleRepository {
	return &ArticleRepository{articleDao: articleDao}
}
