package svc

import (
	"context"
	"fmt"
	"time"

	"codexie.com/w-book-article/internal/config"
	dao "codexie.com/w-book-article/internal/dao/db"
	red "github.com/go-redis/redis"

	"codexie.com/w-book-common/producer"
	"codexie.com/w-book-interact/api/pb/interact"
	"github.com/IBM/sarama"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Cache  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {

	return &ServiceContext{
		Config: c,
	}
}

func CreateCodeRpcClient(c config.Config) interact.InteractionClient {
	return interact.NewInteractionClient(zrpc.MustNewClient(c.InteractRpcConf).Conn())
}

func CreteDbClient(c config.Config) *gorm.DB {
	mysqlConf := c.MySQLConf
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

func CreateRedSync(c config.Config) *redsync.Redsync {
	redisConf := c.RedisConf

	// 创建 Redis 客户端
	rdb := red.NewClient(&red.Options{
		Addr:     redisConf.Host,
		Password: redisConf.Pass,
		DB:       0,
	})
	pool := goredis.NewPool(rdb)

	// 创建 RedSync 实例
	return redsync.New(pool)
}

func CreateRedisClient(c config.Config) *redis.Client {
	redisConf := c.RedisConf
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
	logx.Infof("redis连接成功: %s", pong)
	return myRedis
}

func CreateKafkaProducer(c config.Config) producer.Producer {
	saramaConf := sarama.NewConfig()
	saramaConf.Producer.Return.Successes = true
	saramaConf.Version = sarama.V2_1_0_0
	kafkaProducer, err := sarama.NewSyncProducer(c.KafkaConf.Brokers, saramaConf)
	if err != nil {
		panic(err)
	}
	return producer.NewKafkaProducer(kafkaProducer)
}

func InitTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&dao.Article{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&dao.PublishedArticle{}); err != nil {
		return err
	}
	return nil
}
