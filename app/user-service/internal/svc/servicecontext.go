package svc

import (
	"context"
	"fmt"
	"time"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/model"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type ServiceContext struct {
	Config        config.Config
	DB            *gorm.DB
	Cache         *redis.Client
	CodeRpcClient pb.CodeClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:        c,
		DB:            CreteDbClient(c),
		Cache:         CreateRedisClient(c),
		CodeRpcClient: CreateCodeRpcClient(c),
	}
}

func CreateCodeRpcClient(c config.Config) pb.CodeClient {
	return pb.NewCodeClient(zrpc.MustNewClient(c.CodeRpcConf).Conn())
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

	// auto sync table structure, no need to create table
	if mysqlConf.AutoMigrate {
		if err = InitTables(db); err != nil {
			logx.Errorf("automigrate table failed, codeerr: %v", err)
			panic(err)
		}
	}

	logx.Info("init mysql client instance success.")

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
	logx.Infof("连接成功: %s", pong)
	return myRedis
}

func InitTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return err
	}

	return nil
}
