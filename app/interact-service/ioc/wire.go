//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/dao/cache"
	dao "codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/handler"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/repo"

	"codexie.com/w-book-interact/internal/svc"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/rest"
)

var ServerSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewArticleHandler)

var LogicSet = wire.NewSet(logic.NewArticleLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewAuthorRepository, repo.NewReaderRepository)

var DaoSet = wire.NewSet(dao.NewAuthorDao, dao.NewReaderDao, cache.NewArticleRedis)

var DbSet = wire.NewSet(svc.CreteDbClient, svc.CreateRedisClient)

func NewApp(c config.Config) (*rest.Server, error) {
	panic(wire.Build(
		ServerSet,
		HandlerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DaoSet,
		DbSet,
	))
}
