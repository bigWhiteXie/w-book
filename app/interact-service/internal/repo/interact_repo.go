package repo

import (
	"context"
	"fmt"
	"time"

	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/singleflight"
)

type IInteractRepo interface {
	GetInteraction(ctx context.Context, cntInfo *domain.Interaction) (*domain.Interaction, error)
	AddReadCnt(ctx context.Context, biz string, bizId int64) error
	HandleBatchRead(ctx context.Context, eventBatch []domain.ReadEvent) error
	HandleBatchReadV2(eventBatch []domain.ReadEvent, msgs []*sarama.ConsumerMessage) error

	CreateInteractData(ctx context.Context, readEvt *domain.ReadEvent) error
	GetTopResIdsByLike(ctx context.Context, resourceType string, limit int) ([]int64, error)
	RefreshTopLikeRedis(ctx context.Context, resourceType string, limit int) error
	RefreshTopLikeLocal(ctx context.Context, resourceType string, limit int) error
}

type InteractRepository struct {
	interactDao *db.InteractDao
	recordDao   *db.RecordDao
	cache       cache.InteractCache
	localCache  cache.TopLikeCache
	sg          singleflight.Group
	isTx        bool
}

func NewInteractRepository(readerDao *db.InteractDao, recordDao *db.RecordDao, cache cache.InteractCache, localCache cache.TopLikeCache) IInteractRepo {
	return &InteractRepository{interactDao: readerDao, cache: cache, recordDao: recordDao, localCache: localCache}
}

func (repo *InteractRepository) RefreshTopLikeRedis(ctx context.Context, resourceType string, limit int) error {
	return repo.cache.UpdateRedisZSet(ctx, resourceType, func() ([]*domain.Interaction, error) {
		resources, err := repo.interactDao.GetTopResourcesByLikes("article", limit)
		if err != nil {
			return nil, err
		}
		return fromInteractions(resources), nil
	})
}

func (repo *InteractRepository) RefreshTopLikeLocal(ctx context.Context, resourceType string, limit int) error {
	var (
		resourceIds []int64
		err         error
	)

	if !repo.localCache.TryLock(resourceType) {
		logx.Infof("已有协程在更新资源[%s]的本地缓存,不再重复更新", resourceType)
		return nil
	}
	defer repo.localCache.Unlock(resourceType)
	for i := 0; i <= 3; i++ {
		resourceIds, err = repo.cache.GetTopFromRedisZSet(ctx, resourceType, limit)
		// 指数退避
		if err != nil || len(resourceIds) == 0 {
			if i == 3 {
				logx.Errorf("获取资源[%s]的缓存失败,原因:%s", resourceType, err)
				return err
			}
			time.Sleep(time.Duration((i + 1)) * time.Second)
		}
		break
	}
	return repo.localCache.UpdateResourceCache(resourceType, resourceIds)
}

func (repo *InteractRepository) GetTopResIdsByLike(ctx context.Context, resourceType string, limit int) ([]int64, error) {
	// 先从本地缓存获取
	localRes, err := repo.localCache.GetTopResources(resourceType)
	if err != nil {
		logx.Infof("获取Resource[%s]本地缓存失败:%s", err)
	}
	if len(localRes) >= limit {
		return localRes[:limit], nil
	}

	// 从redis中获取
	resourceIds, err := repo.cache.GetTopFromRedisZSet(ctx, resourceType, 100)
	if len(resourceIds) == 0 || err != nil { // redis中获取不到或失败，走数据库
		entities, err := repo.interactDao.GetTopResourcesByLikes(resourceType, limit)
		if err != nil {
			return nil, err
		}
		for _, entity := range entities {
			resourceIds = append(resourceIds, entity.BizId)
		}

		//更新缓存
		go func() {
			repo.cache.UpdateRedisZSet(ctx, resourceType, func() ([]*domain.Interaction, error) {
				resources, err := repo.interactDao.GetTopResourcesByLikes("article", limit)
				if err != nil {
					return nil, err
				}
				return fromInteractions(resources), nil
			})
		}()
	}
	//更新本地缓存
	repo.localCache.UpdateResourceCache(resourceType, resourceIds)

	return resourceIds, nil
}
func (repo *InteractRepository) GetInteraction(ctx context.Context, cntInfo *domain.Interaction) (*domain.Interaction, error) {
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
	statInfo := result.(*domain.Interaction)
	return statInfo, err
}

func (repo *InteractRepository) HandleBatchReadV2(eventBatch []domain.ReadEvent, msgs []*sarama.ConsumerMessage) error {
	logx.WithContext(context.Background()).Infof("批量阅读事件,长度:%d", len(eventBatch))
	err := repo.HandleBatchRead(context.Background(), eventBatch)
	if err != nil {
		logx.Errorf("处理批量阅读事件失败,原因:%s", err)
	}
	return err
}

func (repo *InteractRepository) AddReadCnt(ctx context.Context, biz string, bizId int64) error {
	repo.interactDao.IncreRead(ctx, biz, bizId)
	return repo.cache.IncreCntIfExist(ctx, fmt.Sprintf(cntInfoKeyFmt, biz, bizId), domain.Read, 1)
}

// HandleBatchRead handles a batch of read events, updating the database and cache accordingly.
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

func fromInteraction(entity *db.Interaction) *domain.Interaction {
	return &domain.Interaction{
		Biz:        entity.Biz,
		BizId:      entity.BizId,
		LikeCnt:    entity.LikeCnt,
		ReadCnt:    entity.ReadCnt,
		CollectCnt: entity.CollectCnt,
	}
}

func fromInteractions(entitis []db.Interaction) []*domain.Interaction {
	res := make([]*domain.Interaction, 0, len(entitis))
	for _, entity := range entitis {
		res = append(res, fromInteraction(&entity))
	}
	return res
}
