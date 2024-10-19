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

func NewReaderRepository(readerDao *db.InteractDao, cache cache.InteractCache) IInteractRepo {
	return &InteractRepository{interactDao: readerDao, cache: cache}
}

func (repo *InteractRepository) FindCntData(ctx context.Context, cntInfo *domain.StatCnt) (*domain.StatCnt, error) {
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
		return repo.interactDao.FindInteractByBiz(ctx, cntInfo.Biz, cntInfo.BizId)
	})
	if err != nil {
		return nil, err
	}
	res := result.(*domain.StatCnt)
	return res, err
}
