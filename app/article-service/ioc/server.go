package ioc

import (
	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-common/middleware"
	"github.com/zeromicro/go-zero/rest"
)

func NewServer(c config.Config, articleHandler *handler.ArticleHandler) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())

	server.Use(middleware.NewJwtMiddleware().Handle)
	handler.RegisterHandlers(server, articleHandler)

	return server
}
