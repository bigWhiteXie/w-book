//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/kafka/producer"
	repo2 "codexie.com/w-book-common/repo"

	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/event"
	"codexie.com/w-book-interact/internal/handler"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/server"

	"codexie.com/w-book-interact/internal/svc"
	"github.com/google/wire"
)

var ServerSet = wire.NewSet(InitServer, InitRpcServer)

var HandlerSet = wire.NewSet(handler.NewInteractHandler, handler.NewCommentHandler, handler.NewMaintainceHandler)

var LogicSet = wire.NewSet(logic.NewInteractLogic, logic.NewCommentLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewCollectRepository, repo.NewInteractRepository, repo.NewLikeInfoRepository, repo2.NewBaseRepo, repo.NewCommentRepo)

var DaoSet = wire.NewSet(cache.NewInteractRedis, cache.NewBigCacheResourceCache, cache.NewCommentRedisCache)

var DbSet = wire.NewSet(ioc.InitGormDB, ioc.InitRedis, ioc.InitRedLock)

var MessageSet = wire.NewSet(ioc.InitKafkaClient, producer.NewKafkaProducer)

var ListenerSet = wire.NewSet(InitConsumers, event.NewCreateEventListener, event.NewBatchReadEventListener, event.NewCommentEvtListener, event.NewCommentLikeEvtListener)

var WokerSet = wire.NewSet(InitCommentJob, InitJobCron)

func NewInteractApp(config config.Config, mysqlConf ioc.MySQLConf, redisConf ioc.RedisConf, kafkaConf ioc.KafkaConf) (*App, error) {
	panic(wire.Build(
		wire.Struct(new(App), "Server", "Consumers", "JobCron"),
		ServerSet,
		HandlerSet,
		WokerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DaoSet,
		DbSet,
		MessageSet,
		ListenerSet,
	))
}

func NewRpcApp(c config.Config, mysqlConf ioc.MySQLConf, redisConf ioc.RedisConf, kafkaConf ioc.KafkaConf) (*server.InteractionServer, error) {
	panic(wire.Build(
		ServerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DaoSet,
		DbSet,
		MessageSet,
	))
}
