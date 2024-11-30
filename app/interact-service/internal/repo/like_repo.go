package repo

import (
	"context"
	"fmt"

	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var (
	cntInfoKeyFmt = "cnt:%s:%d"
)

type ILikeInfoRepository interface {
	Like(ctx context.Context, uid int64, biz string, bizId int64, isLike bool) error
	IsLike(ctx context.Context, uid int64, biz string, bizId int64) (bool, error)
}

type LikeInfoRepository struct {
	interactCache cache.InteractCache
	g             singleflight.Group
	db            *gorm.DB
}

func NewLikeInfoRepository(cache cache.InteractCache, db *gorm.DB) ILikeInfoRepository {
	return &LikeInfoRepository{interactCache: cache, db: db}
}

func (repo *LikeInfoRepository) Like(ctx context.Context, uid int64, biz string, bizId int64, isLike bool) error {
	status := 0
	incre := -1

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		var err error
		InteractDao := db.NewInteractDao(tx)
		likeInfoDao := db.NewLikeInfoDao(tx)
		//查询status
		likeInfo, err := likeInfoDao.FindLikeInfo(ctx, uid, biz, bizId)
		if err != nil {
			return errors.WithMessage(err, "[LikeInfoRepository_] 查询点赞信息失败")
		}
		if likeInfo == nil || likeInfo.Status == 0 {
			status = 1
			incre = 1
		}
		if err = likeInfoDao.UpdateLikeInfo(ctx, uid, biz, bizId, uint8(status)); err != nil {
			return err
		}
		if status > 0 {
			err = InteractDao.IncreLike(ctx, biz, bizId)
		} else {
			err = InteractDao.DecreLike(ctx, biz, bizId)
		}
		return err
	})
	if err != nil {
		return err
	}

	//更新缓存信息
	go func() {
		if err := repo.interactCache.IncreCntIfExist(context.Background(), fmt.Sprintf(cntInfoKeyFmt, biz, bizId), domain.Like, incre); err != nil {
			logx.Errorf("[LikeInfoRepository_Like] 更新资源[%s-%d]点赞缓存失败:%s", biz, bizId, err)
		}
		if err := repo.interactCache.IncrementLikeInZSet(context.Background(), biz, bizId, incre); err != nil {
			logx.Errorf("[LikeInfoRepository_Like] 更新资源[%s-%d]zset失败:%s", biz, bizId, err)
		}
	}()
	return nil
}

func (repo *LikeInfoRepository) IsLike(ctx context.Context, uid int64, biz string, bizId int64) (bool, error) {
	likeDao := db.NewLikeInfoDao(repo.db)
	likeInfo, err := likeDao.FindLikeInfo(ctx, uid, biz, bizId)
	switch {
	case errors.As(err, &gorm.ErrRecordNotFound):
		return false, nil
	case err != nil:
		return false, err
	default:
		return likeInfo.Status == 1, nil
	}
}
