//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-account/internal/config"
	"codexie.com/w-book-account/internal/service"
	"codexie.com/w-book-account/internal/svc"
	"codexie.com/w-book-common/ioc"

	"github.com/google/wire"
)

var AppSet = wire.NewSet(InitRpcServer)

var LogicSet = wire.NewSet(service.NewAccountService)

var SvcSet = wire.NewSet(svc.NewServiceContext)

// var RepoSet = wire.NewSet(repo.NewAccountRepository)

var DbSet = wire.NewSet(InitDB, ioc.InitRedis)

var ConfSet = wire.NewSet(config.GetMySQLConf)

// var MessageSet = wire.NewSet(ioc.InitKafkaClient, producer.NewKafkaProducer)

// var ListenerSet = wire.NewSet(InitConsumers, event.NewCreateEventListener, event.NewBatchReadEventListener)

// var WokerSet = wire.NewSet(worker.NewTopLikeWorker)

func NewPaymentApp(config config.Config) (*App, error) {
	panic(wire.Build(
		wire.Struct(new(App), "RpcServer", "SvcCtx"),
		AppSet,
		LogicSet,
		SvcSet,
		// RepoSet,
		DbSet,
		ConfSet,
	))
}
