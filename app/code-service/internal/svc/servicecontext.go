package svc

import (
	"codexie.com/w-book-code/internal/model"
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"time"

	"codexie.com/w-book-code/internal/config"
	"github.com/redis/go-redis/v9"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type ServiceContext struct {
	Config        config.Config
	DB            *gorm.DB
	Cache         *redis.Client
	ConsumerGroup sarama.ConsumerGroup
	KafkaProvider sarama.AsyncProducer
}

func NewServiceContext(c config.Config) *ServiceContext {
	creteDbClient(c.MySQLConf)
	return &ServiceContext{
		Config:        c,
		DB:            creteDbClient(c.MySQLConf),
		Cache:         createRedisClient(c.RedisConf),
		ConsumerGroup: CreateConsumerGroup(c.KafkaConf),
		KafkaProvider: CreateKafkaProvider(c.KafkaConf),
	}
}

func creteDbClient(mysqlConf config.MySQLConf) *gorm.DB {
	datasource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		mysqlConf.User,
		mysqlConf.Password,
		mysqlConf.Host,
		mysqlConf.Port,
		mysqlConf.Database,
		mysqlConf.CharSet,
		mysqlConf.ParseTime,
		mysqlConf.TimeZone)

	db, err := gorm.Open(mysql.Open(datasource), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   mysqlConf.Gorm.TablePrefix,   // such as: prefix_tableName
			SingularTable: mysqlConf.Gorm.SingularTable, // such as zero_user, not zero_users
		},
	})
	if err != nil {
		panic(err)
	}

	logx.Info("init mysql client instance success.")
	InitTables(db)
	sqlDB, err := db.DB()
	if err != nil {
		logx.Errorf("mysql set connection pool failed, codeerr: %v.", err)
		panic(err)
	}
	sqlDB.SetMaxOpenConns(mysqlConf.Gorm.MaxOpenConns)
	sqlDB.SetMaxIdleConns(mysqlConf.Gorm.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(mysqlConf.Gorm.ConnMaxLifetime) * time.Second)

	return db
}

func createRedisClient(redisConf config.RedisConf) *redis.Client {
	myRedis := redis.NewClient(&redis.Options{
		Addr:     redisConf.Host,
		Password: redisConf.Pass,
		DB:       0,
	})

	// 测试连接
	pong, err := myRedis.Ping(context.Background()).Result()
	if err != nil {
		logx.Errorf("无法连接到 Redis: %s", err)
		return nil
	}
	logx.Infof("连接成功: %s", pong)
	return myRedis
}

func CreateConsumerGroup(conf config.KafkaConf) sarama.ConsumerGroup {
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V2_1_0_0
	client, err := sarama.NewConsumerGroup(conf.Brokers, conf.Topic, saramaConf)
	if err != nil {
		panic(fmt.Sprintf("unable to create kafka consumer group, cause:%s", err))
	}
	return client
}

func CreateKafkaProvider(conf config.KafkaConf) sarama.AsyncProducer {
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V2_1_0_0
	producer, err := sarama.NewAsyncProducer(conf.Brokers, saramaConf)
	if err != nil {
		panic(fmt.Sprintf("unable to create kafka producer, cause:%s", err))
	}
	return producer
}

func InitTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.SmsSendRecord{}); err != nil {
		return err
	}
	return nil
}
