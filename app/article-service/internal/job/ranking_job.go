package job

import (
	"context"

	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-common/job"
)

var (
	defaultTimeExper = "*/1 * * * *"
)

type RankingJob struct {
	rankingLogic *logic.RankingLogic
	timeExper    string
}

func NewRankingJob(rank *logic.RankingLogic, opts ...job.Option) *RankingJob {
	rankJob := &RankingJob{
		rankingLogic: rank,
		timeExper:    defaultTimeExper,
	}
	for _, opt := range opts {
		opt(rankJob)
	}

	return rankJob
}

func (job *RankingJob) Run() error {
	return job.rankingLogic.RefreshTopArticle(context.Background())
}

func (job *RankingJob) Name() string {
	return "article_rank_job"
}

func (job *RankingJob) TimeExper() string {
	return job.timeExper
}
