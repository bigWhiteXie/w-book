package svc

import (
	"context"
	"fmt"
	"time"

	"codexie.com/w-book-code/internal/model"
	"github.com/IBM/sarama"

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
	KafkaProvider sarama.SyncProducer
}

func NewServiceContext(c config.Config) *ServiceContext {
	creteDbClient(c)
	return &ServiceContext{
		Config:        c,
		DB:            creteDbClient(c),
		Cache:         createRedisClient(c),
		ConsumerGroup: CreateSmsConsumerGroup(c),
		KafkaProvider: CreateKafkaProvider(c),
	}
}

func creteDbClient(c config.Config) *gorm.DB {
	datasource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.MySQLConf.User,
		c.MySQLConf.Password,
		c.MySQLConf.Host,
		c.MySQLConf.Port,
		c.MySQLConf.Database,
		c.MySQLConf.CharSet,
		c.MySQLConf.ParseTime,
		c.MySQLConf.TimeZone)

	db, err := gorm.Open(mysql.Open(datasource), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   c.MySQLConf.Gorm.TablePrefix,   // such as: prefix_tableName
			SingularTable: c.MySQLConf.Gorm.SingularTable, // such as zero_user, not zero_users
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
	sqlDB.SetMaxOpenConns(c.MySQLConf.Gorm.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.MySQLConf.Gorm.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(c.MySQLConf.Gorm.ConnMaxLifetime) * time.Second)

	return db
}

func createRedisClient(c config.Config) *redis.Client {
	myRedis := redis.NewClient(&redis.Options{
		Addr:     c.RedisConf.Host,
		Password: c.RedisConf.Pass,
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

func CreateSmsConsumerGroup(c config.Config) sarama.ConsumerGroup {
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V2_1_0_0
	client, err := sarama.NewConsumerGroup(c.KafkaConf.Brokers, "sms-topic", saramaConf)
	if err != nil {
		panic(fmt.Sprintf("unable to create kafka consumer group, cause:%s", err))
	}
	return client
}

func CreateKafkaProvider(c config.Config) sarama.SyncProducer {
	saramaConf := sarama.NewConfig()
	saramaConf.Producer.Return.Successes = true
	saramaConf.Version = sarama.V2_1_0_0
	producer, err := sarama.NewSyncProducer(c.KafkaConf.Brokers, saramaConf)
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
