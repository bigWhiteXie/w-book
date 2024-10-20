package logic

import (
	"context"
	"errors"

	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/types"
)

type InteractLogic struct {
	likeRepo     repo.ILikeInfoRepository
	interactRepo repo.IInteractRepo
	colRepo      repo.ICollectRepository
}

func NewInteractLogic(authorRepo repo.ILikeInfoRepository, readerRepo repo.IInteractRepo, colRepo repo.ICollectRepository) *InteractLogic {
	return &InteractLogic{likeRepo: authorRepo, interactRepo: readerRepo, colRepo: colRepo}
}

// 点赞/取消点赞资源
func (l *InteractLogic) Like(ctx context.Context, req *types.LikeResourceReq) error {
	uid := ctx.Value("id").(int64)
	err := l.likeRepo.Like(ctx, uid, req.Biz, req.BizId, req.Action == 1)

	return err
}

// 收藏/取消收藏资源
func (l *InteractLogic) Collect(ctx context.Context, req *types.CollectResourceReq) error {
	uid := ctx.Value("id").(int64)
	err := l.colRepo.AddCollectionItem(ctx, req.ToDomain(uid))

	return err
}

// 添加/删除收藏夹
func (l *InteractLogic) AddOrDelCollection(ctx context.Context, req *types.CollectionReq) error {
	uid := ctx.Value("id").(int64)
	col := &domain.Collection{
		Id:   req.Id,
		Name: req.Name,
		Uid:  uid,
	}
	if req.Action > 0 {
		return l.colRepo.AddCollection(ctx, col)
	}
	if col.Id == 0 {
		return errors.New("删除收藏夹但没指定id")
	}

	return l.colRepo.DelCollection(ctx, uid, col.Id)
}

// 查询用户的所有收藏夹

// 查询收藏夹下的所有资源
// 查询资源计数信息
func (l *InteractLogic) QueryStatInfo(ctx context.Context, req *types.LikeResourceReq) (*domain.StatCnt, error) {
	statInfo := &domain.StatCnt{
		Biz:   req.Biz,
		BizId: req.BizId,
	}
	cntData, err := l.interactRepo.FindCntData(ctx, statInfo)

	return cntData, err
}
