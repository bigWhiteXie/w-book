package logic

import (
	"context"

	"codexie.com/w-book-article/internal/dao"
	"codexie.com/w-book-article/internal/domain"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-article/internal/types"
)

type ArticleLogic struct {
	authorRepo repo.IAuthorRepository
	readerRepo repo.IReaderRepository
}

func NewArticleLogic(authorRepo repo.IAuthorRepository, readerRepo repo.IReaderRepository) *ArticleLogic {
	return &ArticleLogic{authorRepo: authorRepo, readerRepo: readerRepo}
}

func (l *ArticleLogic) Edit(ctx context.Context, req *types.EditArticleReq) (int64, error) {
	id := ctx.Value("id").(int)
	artDomain := &domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: int64(id),
		},
	}
	return l.authorRepo.Save(ctx, artDomain)
}

func (l *ArticleLogic) Publish(ctx context.Context, req *types.EditArticleReq) (int64, error) {
	id := ctx.Value("id").(int)
	artDomain := &domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: int64(id),
		},
	}
	//保存到制作库
	artId, err := l.authorRepo.Save(ctx, artDomain)
	if err != nil {
		return 0, err
	}

	artDomain.Id = artId
	artDomain.Status = dao.ArticleStatusPublished
	return l.readerRepo.Save(ctx, artDomain)
}
