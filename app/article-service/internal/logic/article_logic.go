package logic

import (
	"context"

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
	id := ctx.Value("id").(int64)
	artDomain := &domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: int64(id),
		},
	}
	//保存到制作库,并刷新缓存
	artId, err := l.authorRepo.Save(ctx, artDomain)
	if err != nil {
		return 0, err
	}

	artDomain.Id = artId
	artDomain.Status = domain.ArticlePublishedStatus
	for i := 0; i < 3; i++ {
		if id, err = l.readerRepo.Save(ctx, artDomain); err == nil {
			return id, err
		}
	}

	return 0, err
}

func (l *ArticleLogic) Page(ctx context.Context, req *types.ArticlePageReq) ([]*domain.Article, error) {
	id := int64(ctx.Value("id").(int))
	return l.authorRepo.SelectPage(ctx, id, req.Page, req.Size)
}

func (l *ArticleLogic) ViewArticle(ctx context.Context, req *types.ArticleViewReq) (*domain.Article, error) {
	if req.IsPublished {
		return l.readerRepo.FindById(ctx, req.Id)
	}
	return l.authorRepo.FindArticleById(ctx, req.Id)
}
