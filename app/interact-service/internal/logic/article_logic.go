package logic

import (
	"context"

	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/types"
)

type InteractLogic struct {
	likeRepo     repo.ILikeInfoRepository
	interactRepo repo.IInteractRepo
}

func NewArticleLogic(authorRepo repo.ILikeInfoRepository, readerRepo repo.IInteractRepo) *InteractLogic {
	return &InteractLogic{likeRepo: authorRepo, interactRepo: readerRepo}
}

func (l *InteractLogic) Like(ctx context.Context, req *types.LikeResourceReq) error {
	uid := ctx.Value("id").(int64)
	err := l.likeRepo.Like(ctx, uid, req.Biz, req.BizId, req.Liked == 1)

	return err
}

func (l *InteractLogic) QueryStatInfo(ctx context.Context, req *types.LikeResourceReq) (*domain.StatCnt, error) {
	statInfo := &domain.StatCnt{
		Biz:   req.Biz,
		BizId: req.BizId,
	}
	cntData, err := l.interactRepo.FindCntData(ctx, statInfo)

	return cntData, err
}
