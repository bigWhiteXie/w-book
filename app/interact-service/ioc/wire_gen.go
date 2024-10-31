// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/handler"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/svc"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/rest"
)

// Injectors from wire.go:

func NewApp(c config.Config) (*rest.Server, error) {
	serviceContext := svc.NewServiceContext(c)
	gormDB := svc.CreteDbClient(c)
	authorDao := db.NewInteractDao(gormDB)
	client := svc.CreateRedisClient(c)
	articleCache := cache.NewInteractRedis(client)
	iAuthorRepository := repo.NewLikeInfoRepository(authorDao, articleCache)
	readerDao := db.NewLikeInfoDao(gormDB)
	iReaderRepository := repo.NewInteractRepository(readerDao, articleCache)
	articleLogic := logic.NewInteractLogic(iAuthorRepository, iReaderRepository)
	articleHandler := handler.NewInteractHandler(serviceContext, articleLogic)
	server := NewServer(c, articleHandler)
	return server, nil
}

// wire.go:

var ServerSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewInteractHandler)

var LogicSet = wire.NewSet(logic.NewInteractLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewLikeInfoRepository, repo.NewInteractRepository)

var DaoSet = wire.NewSet(db.NewInteractDao, db.NewLikeInfoDao, cache.NewInteractRedis)

var DbSet = wire.NewSet(svc.CreteDbClient, svc.CreateRedisClient)
