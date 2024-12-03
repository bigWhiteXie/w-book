//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/dao/cache"
	dao "codexie.com/w-book-article/internal/dao/db"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-article/internal/job"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/kafka/producer"

	"codexie.com/w-book-article/internal/svc"
	"github.com/google/wire"
	"github.com/robfig/cron/v3"
)

var AppSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewArticleHandler)

var LogicSet = wire.NewSet(logic.NewArticleLogic, logic.NewRankingLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewAuthorRepository, repo.NewReaderRepository, repo.NewRankRepo)

var DaoSet = wire.NewSet(dao.NewAuthorDao, dao.NewReaderDao)

var CacheSet = wire.NewSet(cache.NewRankCacheRedis, cache.NewLocalArtTopCache, cache.NewArticleRedis)

var DbSet = wire.NewSet(ioc.InitGormDB, ioc.InitRedis, ioc.InitRedLock)

var MessageSet = wire.NewSet(ioc.InitKafkaClient, producer.NewKafkaProducer)

var RpcSet = wire.NewSet(svc.CreateCodeRpcClient)

var JobSet = wire.NewSet(InitJobStarter, job.NewRankingJob)

func NewApp(cron *cron.Cron, config config.Config, mysqlConf ioc.MySQLConf, redisConf ioc.RedisConf, kafkaConf ioc.KafkaConf) (*App, error) {
	panic(wire.Build(
		AppSet,
		HandlerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DaoSet,
		CacheSet,
		DbSet,
		MessageSet,
		RpcSet,
		JobSet,
	))
}
