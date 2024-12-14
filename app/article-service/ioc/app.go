package ioc

import (
	"time"

	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/handler"
	artJob "codexie.com/w-book-article/internal/job"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-common/job"

	"codexie.com/w-book-common/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/rest"
)

type App struct {
	Server *rest.Server
}

func InitArticleApp(c config.Config, articleHandler *handler.ArticleHandler, redisClient *redis.Client) *App {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, articleHandler)
	return &App{
		Server: server,
	}
}

func InitJobBuilder(rankingJob *artJob.RankingJob, redisClient *redis.Client) *job.JobBuilder {
	cron := cron.New()
	jb := job.NewJobBuilder(cron, redisClient, "article", 60*time.Second)
	jb.AddJob(rankingJob, true)
	return jb
}

func InitRankingJob(ranklogic *logic.RankingLogic) *artJob.RankingJob {
	return artJob.NewRankingJob(ranklogic)
}
