package repo

import (
	"context"
	"fmt"

	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/singleflight"
)

type IInteractRepo interface {
	FindCntData(ctx context.Context, cntInfo *domain.StatCnt) (*domain.StatCnt, error)
	AddReadCnt(ctx context.Context, biz string, bizId int64) error
	HandleBatchRead(ctx context.Context, eventBatch []domain.ReadEvent) error
	CreateInteractData(ctx context.Context, readEvt *domain.ReadEvent) error
}

type InteractRepository struct {
	interactDao *db.InteractDao
	recordDao   *db.RecordDao
	cache       cache.InteractCache
	sg          singleflight.Group
	isTx        bool
}

func NewInteractRepository(readerDao *db.InteractDao, recordDao *db.RecordDao, cache cache.InteractCache) IInteractRepo {
	return &InteractRepository{interactDao: readerDao, cache: cache, recordDao: recordDao}
}

func (repo *InteractRepository) FindCntData(ctx context.Context, cntInfo *domain.StatCnt) (*domain.StatCnt, error) {
	logger := logx.WithContext(ctx)
	key := fmt.Sprintf(cntInfoKeyFmt, cntInfo.Biz, cntInfo.BizId)
	//查询缓存
	info, err := repo.cache.GetStatCnt(ctx, key)
	if err != nil {
		logx.Errorf("查询cnt_stat缓存失败,原因:%s", err)
		return nil, err
	}

	if info != nil {
		return info, nil
	}

	logx.Infof("cnt_stat的缓存信息不存在")

	result, err, _ := repo.sg.Do(key, func() (interface{}, error) {
		entity, err := repo.interactDao.FindInteractByBiz(ctx, cntInfo.Biz, cntInfo.BizId)
		if err != nil {
			return nil, err
		}
		cntStat := fromInteraction(entity)
		if err := repo.cache.CacheStatCnt(ctx, key, cntStat); err != nil {
			logger.Errorf("stat_info数据查询DB成功但缓存失败,原因:%s", err)
		}
		return cntStat, nil
	})
	if err != nil {
		return nil, err
	}
	statInfo := result.(*domain.StatCnt)
	return statInfo, err
}

func (repo *InteractRepository) AddReadCnt(ctx context.Context, biz string, bizId int64) error {
	repo.interactDao.IncreRead(ctx, biz, bizId)
	return repo.cache.IncreCntIfExist(ctx, fmt.Sprintf(cntInfoKeyFmt, biz, bizId), domain.Read, 1)
}

func (repo *InteractRepository) HandleBatchRead(ctx context.Context, eventBatch []domain.ReadEvent) error {
	bizs := make([]string, 0, len(eventBatch))
	bizIds := make([]int64, 0, len(eventBatch))
	uIds := make([]int64, 0, len(eventBatch))

	for _, evt := range eventBatch {
		bizs = append(bizs, evt.Biz)
		bizIds = append(bizIds, evt.BizId)
		uIds = append(uIds, evt.Uid)
	}
	if err := repo.interactDao.BatchIncreRead(ctx, bizs, bizIds); err != nil {
		return err
	}

	if err := repo.recordDao.AddRecords(ctx, bizs, bizIds, uIds); err != nil {
		return err
	}

	go func() {
		for i, biz := range bizs {
			bid := bizIds[i]
			if err := repo.cache.IncreCntIfExist(ctx, fmt.Sprintf(cntInfoKeyFmt, biz, bid), domain.Read, 1); err != nil {
				logx.Errorf("增加缓存阅读数失败,Biz:%s, BizId: %d", biz, bid)
			}
		}
	}()
	return nil
}

func (repo *InteractRepository) CreateInteractData(ctx context.Context, readEvt *domain.ReadEvent) error {
	var (
		entity *db.Interaction
		err    error
	)
	if entity, err = repo.interactDao.CrateCntData(ctx, readEvt.Biz, readEvt.BizId); err != nil {
		return err
	}

	return repo.cache.CacheStatCnt(ctx, fmt.Sprintf(cntInfoKeyFmt, readEvt.Biz, readEvt.BizId), fromInteraction(entity))
}

func fromInteraction(entity *db.Interaction) *domain.StatCnt {
	return &domain.StatCnt{
		Biz:        entity.Biz,
		BizId:      entity.Id,
		LikeCnt:    entity.LikeCnt,
		ReadCnt:    entity.ReadCnt,
		CollectCnt: entity.CollectCnt,
	}
}
