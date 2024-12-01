package repo

import (
	"context"

	"codexie.com/w-book-article/internal/dao/cache"
	"codexie.com/w-book-article/internal/domain"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type IRankRepo interface {
	//刷新缓存
	FreshTopNArticles(ctx context.Context, arts []*domain.Article) error

	//从缓存中得到TopN数据
	GetTopNArticles(ctx context.Context) ([]*domain.Article, error)
}

type RankRepo struct {
	localCache *cache.LocalArtTopCache
	redisCache *cache.RedisArtTopNCache
}

func NewRankRepo(lc *cache.LocalArtTopCache, rc *cache.RedisArtTopNCache) *RankRepo {
	return &RankRepo{
		localCache: lc,
		redisCache: rc,
	}
}

func (repo *RankRepo) FreshTopNArticles(ctx context.Context, arts []*domain.Article) error {
	log := logx.WithContext(ctx)
	log.Infof("[RankRepo] =============刷新缓存==================")
	//更新本地缓存，肯定不会出错
	if err := repo.localCache.ReplaceTopN(ctx, arts); err != nil {
		log.Errorf("[RankRepo] 刷新本地缓存失败：%s", err)
		return errors.WithMessage(err, "[RankRepo] 刷新本地缓存失败")
	}
	if err := repo.redisCache.ReplaceTopN(ctx, arts); err != nil {
		return errors.WithMessage(err, "[RankRepo] 刷新redis缓存失败")
	}

	return nil
}

func (repo *RankRepo) GetTopNArticles(ctx context.Context) ([]*domain.Article, error) {
	log := logx.WithContext(ctx)
	arts, err := repo.localCache.TakeTopNArticles(ctx)
	if err == nil && len(arts) != 0 {
		return arts, nil
	}
	if err != nil {
		log.Errorf("[RankRepo_GetTopNArticles] 从本地缓存中取数据失败:%s", err)
	}

	if arts, err = repo.redisCache.TakeTopNArticles(ctx); err != nil {
		log.Errorf("[RankRepo_GetTopNArticles] 从redis中取数据失败:%s", err)
		return nil, err
	}
	log.Info("[RankRepo_GetTopNArticles] 从redis中取数据到本地缓存")

	return arts, nil
}
