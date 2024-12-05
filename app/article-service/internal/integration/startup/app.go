package startup

import (
	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-common/job"
	"codexie.com/w-book-common/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
)

type App struct {
	Server     *rest.Server
	JobStarter *job.JobBuilder
}

func NewArticleApp(c config.Config, articleHandler *handler.ArticleHandler, redisClient *redis.Client) *App {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.TestUidHandle)
	handler.RegisterHandlers(server, articleHandler)
	return &App{
		Server: server,
	}
}
