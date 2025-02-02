//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-reward/internal/config"
	"codexie.com/w-book-reward/internal/handler"
	"codexie.com/w-book-reward/internal/logic"
	"codexie.com/w-book-reward/internal/repo"
	"codexie.com/w-book-reward/internal/svc"

	"github.com/google/wire"
)

var ServerSet = wire.NewSet(InitServer)

var HandlerSet = wire.NewSet(handler.NewRewardHandler)

var LogicSet = wire.NewSet(logic.NewRewardLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewRewardRepository)

var DbSet = wire.NewSet(ioc.InitGormDB, ioc.InitRedis, ioc.InitRedLock)

// var MessageSet = wire.NewSet(ioc.InitKafkaClient, producer.NewKafkaProducer)

var RpcSet = wire.NewSet(InitPaymentRpcClient)

var ConfSet = wire.NewSet(config.GetKafkaConf, config.GetMySQLConf, config.GetPaymentRpcConf, config.GetRedisConf)

func NewApp(config config.Config) (*App, error) {
	panic(wire.Build(
		wire.Struct(new(App), "Server"),
		ServerSet,
		HandlerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DbSet,
		RpcSet,
		ConfSet,
	))
}
