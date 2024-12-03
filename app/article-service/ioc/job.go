package ioc

import (
	artJob "codexie.com/w-book-article/internal/job"
	"codexie.com/w-book-common/job"
	"github.com/robfig/cron/v3"
)

func InitJobStarter(cron *cron.Cron, rankingJob *artJob.RankingJob) *job.JobBuilder {
	jb := job.NewJobBuilder(cron)
	jb.AddJob(rankingJob, true)
	return jb
}
