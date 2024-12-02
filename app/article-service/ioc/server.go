package ioc

import (
	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-article/internal/job"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func NewServer(c config.Config, articleHandler *handler.ArticleHandler, redisClient *redis.Client, jobStart *job.JobBuilder) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())

	// server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, articleHandler)
	logx.Infof("=========jobStart=============")
	jobStart.Start()
	return server
}

func NewJobBuilder(c config.Config, rankJob *job.RankingJob) *job.JobBuilder {
	jbd := job.InitJobBuilder(rankJob)
	return jbd
}
