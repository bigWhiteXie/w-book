package svc

import (
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/repo/dao"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Cache  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	creteDbClient(c.MySQLConf)
	return &ServiceContext{
		Config: c,
		DB:     creteDbClient(c.MySQLConf),
		Cache:  createRedisClient(c.RedisConf),
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

	// auto sync table structure, no need to create table
	if mysqlConf.AutoMigrate {
		if err = dao.InitTables(db); err != nil {
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