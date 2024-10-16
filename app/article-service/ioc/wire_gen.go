// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/svc"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/rest"
)

// Injectors from wire.go:

func NewApp(c config.Config) (*rest.Server, error) {
	serviceContext := svc.NewServiceContext(c)
	articleLogic := logic.NewArticleLogic()
	articleHandler := handler.NewArticleHandler(serviceContext, articleLogic)
	server := NewServer(c, articleHandler)
	return server, nil
}

// wire.go:

var ServerSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewArticleHandler)

var LogicSet = wire.NewSet(logic.NewArticleLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)
