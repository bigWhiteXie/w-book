//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/handler"
	"codexie.com/w-book-user/internal/logic"
	"codexie.com/w-book-user/internal/repo"
	"codexie.com/w-book-user/internal/repo/cache"
	"codexie.com/w-book-user/internal/repo/dao"
	"codexie.com/w-book-user/internal/svc"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/rest"
)

var ServerSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewUserHandler)
var LogicSet = wire.NewSet(logic.NewUserLogic)
var CtxSet = wire.NewSet(svc.NewServiceContext)
var DBSet = wire.NewSet(repo.NewUserRepository, dao.NewUserDao, cache.NewRedisUserCache)
var BaseSet = wire.NewSet(svc.CreteDbClient, svc.CreateRedisClient, svc.CreateCodeRpcClient)

func NewApp(c config.Config) (*rest.Server, error) {
	panic(wire.Build(
		ServerSet,
		HandlerSet,
		LogicSet,
		DBSet,
		BaseSet,
		// CtxSet,
	))
}
