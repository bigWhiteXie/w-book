package repo

import (
	"context"

	"codexie.com/w-book-common/repo"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/domain"
	"golang.org/x/sync/singleflight"
)

type IRecordRepository interface {
	AddRecord(ctx context.Context, readEvt *domain.ReadEvent) error
	PageByResource(ctx context.Context, biz string, bizId int64, limit, offset int) ([]*domain.Record, error)
	PageByUid(ctx context.Context, uid, limit, offset int) ([]*domain.Record, error)
}

type RecordRepository struct {
	*repo.BaseRepo

	interactCache cache.InteractCache
	g             singleflight.Group
}

func NewRecordRepository(cache cache.InteractCache, baseRepo *repo.BaseRepo) IRecordRepository {
	return &RecordRepository{interactCache: cache, BaseRepo: baseRepo}
}

func (repo *RecordRepository) AddRecord(ctx context.Context, event *domain.ReadEvent) error {
	return nil
}
func (repo *RecordRepository) PageByResource(ctx context.Context, biz string, bizId int64, limit, offset int) ([]*domain.Record, error) {
	return nil, nil
}
func (repo *RecordRepository) PageByUid(ctx context.Context, uid, limit, offset int) ([]*domain.Record, error) {
	return nil, nil
}
