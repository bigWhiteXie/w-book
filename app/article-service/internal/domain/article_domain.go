package domain

import "codexie.com/w-book-article/internal/dao"

var (
	ArticleStatusPublished = 1
)

type Article struct {
	Id      int64
	Title   string
	Content string
	Status  int
	Author  Author
	Utime   int64
	Ctime   int64
}

func FromArticle(article *dao.Article) *Article {
	return &Article{
		Id:      article.Id,
		Title:   article.Title,
		Content: article.Content,
		Author: Author{
			Id: article.AuthorId,
		},
	}
}

func FromPublishedArticle(article *dao.PublishedArticle) *Article {
	return &Article{
		Id:      article.Id,
		Title:   article.Title,
		Content: article.Content,
		Status:  article.Status,
		Author: Author{
			Id: article.AuthorId,
		},
	}
}

func (art Article) ToEntity() *dao.Article {
	return &dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}

type Author struct {
	Id   int64
	Name string
}
