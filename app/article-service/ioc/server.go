package ioc

import (
	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-common/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
)

func NewServer(c config.Config, articleHandler *handler.ArticleHandler, redisClient *redis.Client) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())

	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, articleHandler)

	return server
}
