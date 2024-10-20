// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/dao"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-article/internal/svc"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/rest"
)

// Injectors from wire.go:

func NewApp(c config.Config) (*rest.Server, error) {
	serviceContext := svc.NewServiceContext(c)
	db := svc.CreteDbClient(c)
	authorDao := dao.NewAuthorDao(db)
	iAuthorRepository := repo.NewAuthorRepository(authorDao)
	readerDao := dao.NewReaderDao(db)
	iReaderRepository := repo.NewReaderRepository(readerDao)
	articleLogic := logic.NewArticleLogic(iAuthorRepository, iReaderRepository)
	articleHandler := handler.NewArticleHandler(serviceContext, articleLogic)
	server := NewServer(c, articleHandler)
	return server, nil
}

// wire.go:

var ServerSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewArticleHandler)

var LogicSet = wire.NewSet(logic.NewArticleLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewAuthorRepository, repo.NewReaderRepository)

var DaoSet = wire.NewSet(dao.NewAuthorDao, dao.NewReaderDao)

var DbSet = wire.NewSet(svc.CreteDbClient)
