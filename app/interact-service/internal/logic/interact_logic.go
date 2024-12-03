package logic

import (
	"context"

	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/types"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
)

type InteractLogic struct {
	likeRepo     repo.ILikeInfoRepository
	interactRepo repo.IInteractRepo
	colRepo      repo.ICollectRepository
}

func NewInteractLogic(authorRepo repo.ILikeInfoRepository, readerRepo repo.IInteractRepo, colRepo repo.ICollectRepository) *InteractLogic {
	logic := &InteractLogic{likeRepo: authorRepo, interactRepo: readerRepo, colRepo: colRepo}
	return logic
}

// 点赞/取消点赞资源
func (l *InteractLogic) Like(ctx context.Context, req *types.OpResourceReq) error {
	uid := int64(ctx.Value("id").(int))
	err := l.likeRepo.Like(ctx, uid, req.Biz, int64(req.BizId))

	return err
}

// 收藏/取消收藏资源
func (l *InteractLogic) Collect(ctx context.Context, req *types.CollectResourceReq) error {
	uid := int64(ctx.Value("id").(int))
	err := l.colRepo.AddCollectionItem(ctx, req.ToDomain(uid))

	return err
}

func (l *InteractLogic) AddRead(ctx context.Context, biz string, bizId int64) error {
	return l.interactRepo.AddReadCnt(ctx, biz, bizId)
}

// 查询用户的所有收藏夹
func (l *InteractLogic) QueryInteractionInfos(ctx context.Context, biz string, bizIds []int64) ([]*domain.Interaction, error) {
	return l.interactRepo.GetInteractions(ctx, biz, bizIds)
}

// 添加/删除收藏夹
func (l *InteractLogic) AddOrDelCollection(ctx context.Context, req *types.CollectionReq) error {
	uid := int64(ctx.Value("id").(int))
	col := &domain.Collection{
		Id:   req.Id,
		Name: req.Name,
		Uid:  uid,
	}
	if req.Action > 0 {
		return l.colRepo.AddCollection(ctx, col)
	}
	if col.Id == 0 {
		return errors.Wrap(codeerr.WithCode(codeerr.SystemErrCode, "[InteractLogic_AddOrDelCollection] 删除收藏夹id不能为空"), "")
	}

	return l.colRepo.DelCollection(ctx, uid, col.Id)
}

// 查询用户的所有收藏夹

// 查询收藏夹下的所有资源
// 查询资源计数信息
func (l *InteractLogic) QueryStatInfo(ctx context.Context, req *types.OpResourceReq) (*domain.Interaction, error) {
	var (
		isLiked     bool
		isCollected bool
		statInfo    = &domain.Interaction{
			Biz:   req.Biz,
			BizId: int64(req.BizId),
		}
	)

	eg, _ := errgroup.WithContext(ctx)
	// 获取点赞数、阅读数、收藏数等统计信息
	eg.Go(func() error {
		var err error
		statInfo, err = l.interactRepo.GetInteraction(ctx, statInfo)
		logx.Errorf("[InteractLogic_QueryStatInfo] 获取统计信息 err:%s", errors.Cause(err))
		return err
	})

	//判断是否点赞
	eg.Go(func() error {
		var err error
		isLiked, err = l.likeRepo.IsLike(ctx, req.Uid, req.Biz, req.BizId)
		logx.Errorf("[InteractLogic_QueryStatInfo] 获取统计信息 err:%s", errors.Cause(err))
		return err
	})

	//判断是否收藏
	eg.Go(func() error {
		var err error
		isCollected, err = l.colRepo.IsCollected(ctx, req.Uid, req.Biz, req.BizId)
		logx.Errorf("[InteractLogic_QueryStatInfo] 获取统计信息 err:%s", errors.Cause(err))
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	statInfo.IsCollected = isCollected
	statInfo.IsLiked = isLiked
	return statInfo, nil
}

func (l *InteractLogic) GetTopLike(ctx context.Context, biz string) ([]int64, error) {
	return l.interactRepo.GetTopResIdsByLike(ctx, biz, 100)
}
