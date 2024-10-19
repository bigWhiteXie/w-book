package ioc

import (
	"codexie.com/w-book-common/middleware"
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/handler"
	"github.com/zeromicro/go-zero/rest"
)

func NewServer(c config.Config, articleHandler *handler.InteractHandler) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())

	server.Use(middleware.NewJwtMiddleware().Handle)
	handler.RegisterHandlers(server, articleHandler)

	return server
}
