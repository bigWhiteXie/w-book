package worker

import (
	"context"
	"sync"
	"time"

	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-common/job"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/repo"
)

var (
	defaultTimeExper = "*/1 * * * *"

	idSize          = 200
	preloadPageSize = 20
	maxConcurrency  = 50 // 最大并发数
	bizTypesKey     = []string{""}
)

// 缓存热点资源评论 定时任务
type CommentJob struct {
	commentLogic *logic.CommentLogic
	resourceRepo repo.IResourceRepo
	timeExper    string
	cache        cache.CommentCache
}

func NewRootCommentJob(rank *logic.CommentLogic, cache cache.CommentCache, opts ...job.Option) *CommentJob {
	rankJob := &CommentJob{
		commentLogic: rank,
		cache:        cache,
		timeExper:    defaultTimeExper,
	}
	for _, opt := range opts {
		opt(rankJob)
	}

	return rankJob
}

func (job *CommentJob) Run() error {
	wg := &sync.WaitGroup{}
	ctx := context.Background()
	for _, bizType := range domain.BizTypes {
		wg.Add(1)
		// 每种资源类似开一个协程
		go func(bizType string) {
			defer wg.Done()
			offset := 0
			for {
				// 从Redis分页查询该类资源热点列表
				ids, err := job.resourceRepo.GetResourceIDs(ctx, bizType, offset, idSize)
				if err != nil {
					codeerr.LogCodeError(ctx, "获取热点资源失败", err, "获取热点资源失败,biz=%s", bizType)
					return
				}

				if len(ids) == 0 {
					break
				}

				// 数据库批量查询这批热点资源的首页评论
				comments, err := job.commentLogic.GetRootCommentsByBizIDs(ctx, bizType, ids, 20)
				if err != nil {
					codeerr.LogCodeError(ctx, "获取热点资源评论失败", err, "获取热点资源评论失败,biz=%s", bizType)
					return
				}

				// 并发将评论缓存到Redis
				for _, bizID := range ids {
					go func(biz string, id int64) {
						if err := job.cache.SetRootComments(ctx, bizType, id, comments, 10*time.Minute); err != nil {
							codeerr.LogCodeError(ctx, "缓存热点资源评论失败", err, "缓存热点资源评论失败,biz=%s,id=%d", bizType, id)
						}
					}(bizType, bizID)
				}

				if len(ids) < idSize {
					break
				}
				offset += idSize
			}
		}(bizType)
	}

	wg.Wait()
	return nil
}

func (job *CommentJob) Name() string {
	return "resource_comment_job"
}

func (job *CommentJob) TimeExper() string {
	return job.timeExper
}
