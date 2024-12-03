package logic

import (
	"context"
	"time"

	"codexie.com/w-book-article/internal/domain"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-common/queue"
	"codexie.com/w-book-interact/api/pb/interact"
	"github.com/go-redsync/redsync/v4"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	batchSize = 2000
	topN      = 100
)

type RankingLogic struct {
	readerRepo  repo.IReaderRepository
	rankRepo    repo.IRankRepo
	interactRpc interact.InteractionClient
	redLock     *redsync.Redsync
}

func NewRankingLogic(readerRepo repo.IReaderRepository, rankRepo *repo.RankRepo, rs *redsync.Redsync, interactRpc interact.InteractionClient) *RankingLogic {
	return &RankingLogic{readerRepo: readerRepo, rankRepo: rankRepo, interactRpc: interactRpc, redLock: rs}
}

// 批量查询文章,计算每个文章的分数(点赞)，添加到topQueue中
func (l *RankingLogic) RankTopNFromDB(ctx context.Context) ([]*domain.Article, error) {
	lastId := int64(0)
	topQueue := queue.NewFixedSizePriorityQueue[*domain.Article](topN)
	for {
		var (
			arts []*domain.Article
			err  error
		)

		if arts, err = l.readerRepo.ListArticlesV2(ctx, lastId, batchSize, false); err != nil {
			return nil, err
		}

		lastId = arts[len(arts)-1].Id
		ids := make([]int64, 0, len(arts))
		for _, art := range arts {
			ids = append(ids, art.Id)
		}
		result, err := l.interactRpc.QueryInteractionsInfo(ctx, &interact.QueryInteractionsReq{
			Biz:    domain.Biz,
			BizIds: ids,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "[RankTopNFromDB] RPC访问交互统计资源失败:%s", err)
		}
		cntMap := make(map[int64]*interact.InteractionResult, len(result.Interactions))
		for _, cnt := range result.Interactions {
			cntMap[cnt.BizId] = cnt
		}
		for _, art := range arts {
			if cnt, ok := cntMap[art.Id]; ok {
				score := int(cnt.LikeCnt)
				art.LikeCnt = cnt.LikeCnt
				art.CollectCnt = cnt.CollectCnt
				art.ReadCnt = cnt.ReadCnt
				err := topQueue.Enqueue(art, score)
				if err != nil {
					logx.Errorf("TopQueue入队失败:%s", err)
				}
			}
		}

		if len(arts) < batchSize {
			break
		}
	}

	return topQueue.GetAll(), nil
}

func (l *RankingLogic) RefreshTopArticle(ctx context.Context) error {
	lockKey := "rank:top:article:lock"
	mutex := l.redLock.NewMutex(lockKey,
		redsync.WithExpiry(60*time.Second),
		redsync.WithTries(1),
	)

	if err := mutex.TryLock(); err != nil {
		if err == redsync.ErrFailed {
			logx.Infof("[RankingLogic_RefreshTopArticle] 当前其它服务正在占用锁:%s", lockKey)
			return nil
		}
		logx.Errorf("获取分布式锁%s失败", lockKey)
		return errors.Wrapf(err, "[RankingLogic_RefreshTopArticle] 获取分布式锁失败,lockKey=%s", lockKey)
	}
	defer mutex.Unlock()

	arts, err := l.RankTopNFromDB(ctx)
	if err != nil {
		return errors.WithMessage(err, "[RefreshTopArticle] 从数据库取数据异常导致刷新热榜article失败")
	}

	return l.rankRepo.FreshTopNArticles(ctx, arts)
}

func (l *RankingLogic) GetTopArticles(ctx context.Context) ([]*domain.Article, error) {
	arts, err := l.rankRepo.GetTopNArticles(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "[RankingLogic] GetTopArticles失败")
	}

	return arts, nil
}
