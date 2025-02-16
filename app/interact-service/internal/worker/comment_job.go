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
					codeerr.LogCodeError(ctx, "获取热点资源失败", "获取热点资源失败,biz=%s", bizType)
					return
				}

				if len(ids) == 0 {
					break
				}

				// 数据库批量查询这批热点资源的首页评论
				comments, err := job.commentLogic.GetRootCommentsByBizIDs(ctx, bizType, ids, 20)
				if err != nil {
					codeerr.LogCodeError(ctx, "获取热点资源评论失败", "获取热点资源评论失败,biz=%s", bizType)
					return
				}
				// 将评论按资源id分组
				commentsMap := make(map[int64][]*domain.Comment)
				for _, comment := range comments {
					commentsMap[comment.BizID] = append(commentsMap[comment.BizID], comment)
				}

				// 每个资源开启一个协程将它相关的首页评论缓存到Redis
				subWg := &sync.WaitGroup{}
				for _, bizID := range ids {
					subWg.Add(1)
					go func(biz string, id int64) {
						defer subWg.Done()
						if err := job.cache.SetRootComments(ctx, bizType, id, commentsMap[id], 10*time.Minute); err != nil {
							codeerr.LogCodeError(ctx, "缓存热点资源评论失败", "缓存热点资源评论失败,biz=%s,id=%d", bizType, id)
							return
						}

						// 查询每条评论的用户点赞情况并缓存到redis中
						for _, comment := range commentsMap[id] {
							userIds, err := job.commentLogic.GetCommentLikeUserIDs(ctx, comment.ID)
							if err != nil {
								return
							}
							job.commentLogic.AddCommentFilter(ctx, comment.ID, userIds, 30*time.Minute)
						}

					}(bizType, bizID)
				}
				subWg.Wait()
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
