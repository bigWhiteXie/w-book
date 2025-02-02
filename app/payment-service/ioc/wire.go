//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/kafka/producer"
	"codexie.com/w-book-payment/internal/config"
	"codexie.com/w-book-payment/internal/handler"
	"codexie.com/w-book-payment/internal/logic"
	"codexie.com/w-book-payment/internal/repo"
	"codexie.com/w-book-payment/internal/svc"

	"github.com/google/wire"
)

var AppSet = wire.NewSet(InitServer, InitRpcServer, InitJobCron)

var HandlerSet = wire.NewSet(handler.NewAliPayHandler)

var LogicSet = wire.NewSet(logic.NewAliPayLogic, logic.NewPaymentLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewPaymentRepository, repo.NewPayMsgRepo)

// var DaoSet = wire.NewSet(cache.NewInteractRedis, cache.NewBigCacheResourceCache)

var DbSet = wire.NewSet(ioc.InitGormDB, ioc.InitRedis)

var JobSet = wire.NewSet(InitPayStatusJob)

var ConfSet = wire.NewSet(config.GetAliPayConfig, config.GetMySQLConf, config.GetRedisConf, config.GetKafkaConf)

var MessageSet = wire.NewSet(ioc.InitKafkaClient, producer.NewKafkaProducer)

// var ListenerSet = wire.NewSet(InitConsumers, event.NewCreateEventListener, event.NewBatchReadEventListener)

// var WokerSet = wire.NewSet(worker.NewTopLikeWorker)

func NewPaymentApp(config config.Config) (*App, error) {
	panic(wire.Build(
		wire.Struct(new(App), "Server", "RpcServer", "SvcCtx", "JobCron"),
		AppSet,
		HandlerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		// DaoSet,
		DbSet,
		JobSet,
		MessageSet,
		// ListenerSet,
		ConfSet,
	))
}
