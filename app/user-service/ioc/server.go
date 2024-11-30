package ioc

import (
	"codexie.com/w-book-common/middleware"
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/handler"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
)

func NewServer(c config.Config, userHandler *handler.UserHandler, client *redis.Client) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())

	server.Use(middleware.NewJwtMiddleware(client).Handle)
	handler.RegisterHandlers(server, c, userHandler)

	return server
}
