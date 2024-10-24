package repo

import (
	"context"
	"errors"
	"fmt"

	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
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
	if isLike {
		status = 1
		incre = 1
	}
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		var err error
		InteractDao := db.NewInteractDao(tx)
		likeInfoDao := db.NewLikeInfoDao(tx)
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
		if err := repo.interactCache.IncreCntIfExist(ctx, fmt.Sprintf(cntInfoKeyFmt, biz, bizId), domain.Like, incre); err != nil {
			logx.Errorf("更新资源[%s-%d]点赞缓存失败", biz, bizId)
			return
		}
		repo.interactCache.IncrementLikeInZSet(ctx, biz, bizId, incre)
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
