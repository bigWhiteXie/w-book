package ioc

import (
	"time"

	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/handler"
	artJob "codexie.com/w-book-article/internal/job"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-common/job"

	middleware "codexie.com/w-book-common/middleware/auth"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/rest"
)

type App struct {
	Server     *rest.Server
	JobStarter *job.JobCron
}

func InitServer(c config.Config, articleHandler *handler.ArticleHandler, redisClient *redis.Client) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, articleHandler)
	return server
}

func InitJobBuilder(rankingJob *artJob.RankingJob, redisClient *redis.Client) *job.JobCron {
	cron := cron.New()
	jb := job.NewJobCron(cron, redisClient, "article", 60*time.Second)
	jb.AddJob(rankingJob, true)
	return jb
}

func InitRankingJob(ranklogic *logic.RankingLogic) *artJob.RankingJob {
	return artJob.NewRankingJob(ranklogic)
}
