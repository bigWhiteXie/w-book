package logic

import (
	"context"

	"codexie.com/w-book-article/internal/domain"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-common/common/queue"
	"codexie.com/w-book-interact/api/pb/interact"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	batchSize = 1000
)

type RankingLogic struct {
	readerRepo  repo.IReaderRepository
	interactRpc interact.InteractionClient
	topQueue    *queue.FixedSizePriorityQueue[*domain.Article]
}

func NewRankingLogic(readerRepo repo.IReaderRepository, interactRpc interact.InteractionClient) *RankingLogic {
	topQueue := queue.NewFixedSizePriorityQueue[*domain.Article](100)
	return &RankingLogic{readerRepo: readerRepo, topQueue: topQueue}
}

// 批量查询文章,计算每个文章的分数(点赞)，添加到topQueue中
func (l *RankingLogic) RankTopN(ctx context.Context) ([]*domain.Article, error) {
	log := logx.WithContext(ctx)
	offset := 0
	for {
		var arts []*domain.Article
		arts, err := l.readerRepo.ListArticles(ctx, offset, batchSize)
		ids := make([]int64, 0, len(arts))

		if err != nil {
			return nil, err
		}

		for _, art := range arts {
			ids = append(ids, art.Id)
		}
		result, err := l.interactRpc.QueryInteractionsInfo(ctx, &interact.QueryInteractionsReq{
			Biz:    "article",
			BizIds: ids,
		})

		if err != nil {
			return nil, err
		}

		for i, art := range arts {
			score := int(result.Interactions[i].LikeCnt)
			err := l.topQueue.Enqueue(art, score)
			if err != nil {
				log.Info("[RankTopN] TopN队满")
			}
			val, minScore, _ := l.topQueue.Dequeue()
			if score > minScore {
				l.topQueue.Enqueue(art, score)
			} else {
				l.topQueue.Enqueue(val, minScore)
			}
		}

		if len(arts) < batchSize {
			break
		}
		offset += len(arts)
	}

	return l.topQueue.GetAll(), nil
}
