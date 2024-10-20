//go:build wireinject
// +build wireinject

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

var ServerSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewArticleHandler)

var LogicSet = wire.NewSet(logic.NewArticleLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewAuthorRepository, repo.NewReaderRepository)

var DaoSet = wire.NewSet(dao.NewAuthorDao, dao.NewReaderDao)

var DbSet = wire.NewSet(svc.CreteDbClient)

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
