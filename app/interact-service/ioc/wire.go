//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/dao/cache"
	dao "codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/event"
	"codexie.com/w-book-interact/internal/handler"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/server"
	"codexie.com/w-book-interact/internal/worker"

	"codexie.com/w-book-interact/internal/svc"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/rest"
)

var ServerSet = wire.NewSet(NewServer, NewRpcServer)

var HandlerSet = wire.NewSet(handler.NewInteractHandler)

var LogicSet = wire.NewSet(logic.NewInteractLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewCollectRepository, repo.NewInteractRepository, repo.NewLikeInfoRepository)

var DaoSet = wire.NewSet(dao.NewCollectionDao, dao.NewInteractDao, dao.NewLikeInfoDao, dao.NewRecordDao, cache.NewInteractRedis, cache.NewBigCacheResourceCache)

var DbSet = wire.NewSet(svc.CreteDbClient, svc.CreateRedisClient, svc.CreateRedSync)

var ListenerSet = wire.NewSet(event.NewBatchReadEventListener, event.NewCreateEventListener)

var WokerSet = wire.NewSet(worker.NewTopLikeWorker)

func NewApp(c config.Config) (*rest.Server, error) {
	panic(wire.Build(
		ServerSet,
		HandlerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DaoSet,
		DbSet,
		WokerSet,
		ListenerSet,
	))
}

func NewRpcApp(c config.Config) (*server.InteractionServer, error) {
	panic(wire.Build(
		ServerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DaoSet,
		DbSet,
	))
}
