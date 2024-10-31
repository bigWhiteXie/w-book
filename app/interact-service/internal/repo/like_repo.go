package repo

import (
	"context"
	"fmt"

	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var (
	cntInfoKeyFmt = "cnt:%s:%d"
)

type ILikeInfoRepository interface {
	Like(ctx context.Context, uid int64, biz string, bizId int64, isLike bool) error
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
	return repo.interactCache.IncreCntIfExist(ctx, fmt.Sprintf(cntInfoKeyFmt, biz, bizId), domain.Like, incre)
}
