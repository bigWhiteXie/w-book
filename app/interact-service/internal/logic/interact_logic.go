package logic

import (
	"context"
	"errors"
	"time"

	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/types"
	"golang.org/x/sync/errgroup"
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
func (l *InteractLogic) Like(ctx context.Context, req *types.OpResourceReq) error {
	uid := int64(ctx.Value("id").(int))
	err := l.likeRepo.Like(ctx, uid, req.Biz, int64(req.BizId), req.Action == 1)

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
		return errors.New("删除收藏夹但没指定id")
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
		return err
	})

	//判断是否点赞
	eg.Go(func() error {
		var err error
		isLiked, err = l.likeRepo.IsLike(ctx, req.Uid, req.Biz, req.BizId)
		return err
	})

	//判断是否收藏
	eg.Go(func() error {
		var err error
		isCollected, err = l.colRepo.IsCollected(ctx, req.Uid, req.Biz, req.BizId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	statInfo.IsCollected = isCollected
	statInfo.IsLiked = isLiked
	return statInfo, nil
}

func (l *InteractLogic) StartPeriodicCacheUpdates(ctx context.Context) {
	ticker1 := time.NewTicker(1 * time.Minute)
	ticker5 := time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-ticker1.C:
				//todo:更新本地缓存
			case <-ticker5.C:
				l.interactRepo.UpdateRedisZSet(ctx, "article", 500)
			case <-ctx.Done():
				ticker1.Stop()
				ticker5.Stop()
				return
			}
		}
	}()
}
