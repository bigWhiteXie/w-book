package repo

import (
	"context"

	"codexie.com/w-book-common/repo"
	"codexie.com/w-book-interact/internal/dao/cache"
	"gorm.io/gorm"
)

type IResourceRepo interface {
	GetResourceIDs(ctx context.Context, biz string, offset, size int) ([]int64, error)
}

type ResourceRepo struct {
	*repo.BaseRepo
	cache cache.ResourceCache
}

func NewResourceRepo(gormDb *gorm.DB, cache cache.ResourceCache) *ResourceRepo {
	return &ResourceRepo{
		BaseRepo: repo.NewBaseRepo(gormDb),
		cache:    cache,
	}
}

func (r *ResourceRepo) GetResourceIDs(ctx context.Context, biz string, offset, size int) ([]int64, error) {
	// 从缓存获取
	ids, err := r.cache.GetResourceIDs(ctx, biz, offset, size)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
