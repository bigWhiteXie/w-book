package job

import (
	"context"

	"codexie.com/w-book-article/internal/logic"
)

type RankingJob struct {
	rankingLogic *logic.RankingLogic
}

func NewRankingJob(rank *logic.RankingLogic) *RankingJob {
	return &RankingJob{
		rankingLogic: rank,
	}
}

func (job *RankingJob) Run() error {
	return job.rankingLogic.RefreshTopArticle(context.Background())
}

func (job *RankingJob) Name() string {
	return "article_rank_job"
}

func (job *RankingJob) TimeExper() string {
	return "*/1 * * * *"
}
