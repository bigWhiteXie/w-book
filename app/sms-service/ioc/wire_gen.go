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

// Injectors from wire.go:

func NewApp(conf config.Config, mysqlConf ioc.MySQLConf, redisConf ioc.RedisConf, kafkaConf ioc.KafkaConf, smsConf sms.SmsConf, labelConf metric.ConstMetricLabelsConf) (*App, error) {
	serviceContext := svc.NewServiceContext(conf)
	client := ioc.InitRedis(redisConf)
	codeRedisCache := cache.NewCodeRedisCache(client)
	db := ioc.InitGormDB(mysqlConf)
	codeDao := dao.NewCodeDao(db)
	smsRepo := repo.NewSmsRepo(codeRedisCache, codeDao)
	saramaClient := ioc.InitKafkaClient(kafkaConf)
	producerProducer := producer.NewKafkaProducer(saramaClient)
	prometheusSmsLogic := InitPrometheusSmsService(smsConf, labelConf)
	aSyncSmsLogic := InitKafkaSmsService(prometheusSmsLogic, smsRepo, producerProducer)
	codeLogic := InitCodeLogic(smsRepo, producerProducer, aSyncSmsLogic)
	smsEvtListener := event.NewSmsEvtListener(saramaClient, smsRepo, prometheusSmsLogic)
	app := NewSmsApp(serviceContext, codeLogic, smsEvtListener)
	return app, nil
}

// wire.go:

var AppSet = wire.NewSet(NewSmsApp)

var LogicSet = wire.NewSet(InitKafkaSmsService, InitCodeLogic, InitPrometheusSmsService)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewSmsRepo)

var DaoSet = wire.NewSet(dao.NewCodeDao)

var CacheSet = wire.NewSet(cache.NewCodeRedisCache)

var DbSet = wire.NewSet(ioc.InitGormDB, ioc.InitRedis)

var MessageSet = wire.NewSet(ioc.InitKafkaClient, producer.NewKafkaProducer, event.NewSmsEvtListener)
