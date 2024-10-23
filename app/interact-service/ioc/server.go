package ioc

import (
	"codexie.com/w-book-common/middleware"
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/event"
	"codexie.com/w-book-interact/internal/handler"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
)

func NewServer(c config.Config, articleHandler *handler.InteractHandler, redisClient *redis.Client, readListener *event.ReadEventListener, createListener *event.CreateEventListener) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())

	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, articleHandler)
	readListener.StartListner()
	createListener.StartListner()
	return server
}
