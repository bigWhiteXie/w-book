package repo

import (
	"context"
	"fmt"
	"time"

	"codexie.com/w-book-common/user"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"golang.org/x/sync/singleflight"
)

type ICollectRepository interface {
	AddCollection(ctx context.Context, col *domain.Collection) error
	AddCollectionItem(ctx context.Context, col *domain.CollectionItem) (err error)
	DelCollection(ctx context.Context, uid, cid int64) error
	IsCollected(ctx context.Context, uid int64, biz string, bizId int64) (bool, error)
}

type CollectRepository struct {
	interactCache cache.InteractCache
	g             singleflight.Group
	collectDao    *db.CollectionDao
}

func NewCollectRepository(cache cache.InteractCache, dao *db.CollectionDao) ICollectRepository {
	return &CollectRepository{interactCache: cache, collectDao: dao}
}

func (repo *CollectRepository) AddCollection(ctx context.Context, col *domain.Collection) error {
	now := time.Now().UnixMilli()
	entity := &db.Collection{
		Name:  col.Name,
		Uid:   col.Uid,
		Count: 0,
		Ctime: now,
		Utime: now,
	}
	return repo.collectDao.AddCollection(ctx, entity)
}

func (repo *CollectRepository) DelCollection(ctx context.Context, uid, cid int64) error {
	return repo.collectDao.DelCollection(ctx, uid, cid)
}

func (repo *CollectRepository) AddCollectionItem(ctx context.Context, col *domain.CollectionItem) (err error) {
	uid := user.GetUidByCtx(ctx)
	//todo: 查询收藏夹和资源的信息，若拿不到则返回error
	now := time.Now().UnixMilli()
	entity := &db.CollectionItem{
		Uid:   uid,
		Biz:   col.Biz,
		BizId: col.BizId,
		Cid:   col.Cid,
		Ctime: now,
		Utime: now,
	}
	incre := -1
	if col.Action == 1 {
		_, err = repo.collectDao.AddCollectionItem(ctx, entity)
		incre = 1
	} else {
		logx.Infof("删除数据 id:%d", col.Id)
		entity.Id = col.Id
		_, err = repo.collectDao.DelCollectionItem(ctx, entity)
	}
	if err != nil {
		return err
	}
	return repo.interactCache.IncreCntIfExist(ctx, fmt.Sprintf(cntInfoKeyFmt, col.Biz, col.BizId), domain.Collect, incre)
}

func (repo *CollectRepository) IsCollected(ctx context.Context, uid int64, biz string, bizId int64) (bool, error) {
	item, err := repo.collectDao.FindCollectionItem(ctx, uid, bizId, biz)
	switch {
	case errors.Cause(err) == gorm.ErrRecordNotFound:
		return false, nil
	case err != nil:
		return false, err
	default:
		return item != nil, nil
	}
}

func FromCollection(entity *db.Collection) *domain.Collection {
	if entity == nil {
		return nil
	}
	return &domain.Collection{
		Id:    entity.Id,
		Name:  entity.Name,
		Uid:   entity.Uid,
		Count: entity.Count,
		Ctime: entity.Ctime,
		Utime: entity.Utime,
	}
}
