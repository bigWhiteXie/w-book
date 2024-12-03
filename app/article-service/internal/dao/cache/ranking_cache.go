package cache

import "codexie.com/w-book-article/internal/domain"

type ArticleRankCache interface {
	ReplaceTopN(arts []*domain.Article) error
	TakeTopN() ([]*domain.Article, error)
}
