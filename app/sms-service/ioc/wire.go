//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-code/internal/config"
	"codexie.com/w-book-code/internal/event"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-code/internal/repo/cache"
	"codexie.com/w-book-code/internal/repo/dao"
	"codexie.com/w-book-code/internal/svc"
	"codexie.com/w-book-code/pkg/sms"
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/kafka/producer"
	"codexie.com/w-book-common/metric"
	"github.com/google/wire"
)

var AppSet = wire.NewSet(NewSmsApp)

var LogicSet = wire.NewSet(InitKafkaSmsService, InitCodeLogic, InitPrometheusSmsService)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewSmsRepo)

var DaoSet = wire.NewSet(dao.NewCodeDao)

var CacheSet = wire.NewSet(cache.NewCodeRedisCache)

var DbSet = wire.NewSet(ioc.InitGormDB, ioc.InitRedis)

var MessageSet = wire.NewSet(ioc.InitKafkaClient, producer.NewKafkaProducer, event.NewSmsEvtListener)

func NewApp(conf config.Config, mysqlConf ioc.MySQLConf, redisConf ioc.RedisConf, kafkaConf ioc.KafkaConf, smsConf sms.SmsConf, labelConf metric.ConstMetricLabelsConf) (*App, error) {
	panic(wire.Build(
		AppSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DaoSet,
		CacheSet,
		DbSet,
		MessageSet,
	))
}
