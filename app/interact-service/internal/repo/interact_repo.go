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
}

type InteractRepository struct {
	interactDao *db.InteractDao
	cache       cache.InteractCache
	sg          singleflight.Group
	isTx        bool
}

func NewInteractRepository(readerDao *db.InteractDao, cache cache.InteractCache) IInteractRepo {
	return &InteractRepository{interactDao: readerDao, cache: cache}
}

func (repo *InteractRepository) FindCntData(ctx context.Context, cntInfo *domain.StatCnt) (*domain.StatCnt, error) {
	logger := logx.WithContext(ctx)
	key := fmt.Sprintf(cntInfoKeyFmt, cntInfo.Biz, cntInfo.BizId)
	//查询缓存
	info, err := repo.cache.GetStatCnt(ctx, key)
	if err != nil {
		logx.Errorf("查询cnt_stat缓存失败,原因:%s", err)
	}
	if info != nil {
		return info, err
	}

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

func fromInteraction(entity *db.Interaction) *domain.StatCnt {
	return &domain.StatCnt{
		Biz:        entity.Biz,
		BizId:      entity.Id,
		LikeCnt:    entity.LikeCnt,
		ReadCnt:    entity.ReadCnt,
		CollectCnt: entity.CollectCnt,
	}
}
