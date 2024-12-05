package ioc

import (
	"time"

	artJob "codexie.com/w-book-article/internal/job"
	"codexie.com/w-book-common/job"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

func InitJobStarter(cron *cron.Cron, rankingJob *artJob.RankingJob, redisClient *redis.Client) *job.JobBuilder {
	jb := job.NewJobBuilder(cron, redisClient, "article", 60*time.Second)
	jb.AddJob(rankingJob, true)
	return jb
}
