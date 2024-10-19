package svc

import (
	"context"
	"fmt"
	"time"

	"codexie.com/w-book-interact/internal/config"
	dao "codexie.com/w-book-interact/internal/dao/db"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
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

func InitTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&dao.Interaction{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&dao.LikeInfo{}); err != nil {
		return err
	}
	return nil
}
