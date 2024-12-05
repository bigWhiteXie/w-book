package startup

import (
	"context"
	"database/sql"
	"time"

	"codexie.com/w-book-article/internal/dao/db"
	"github.com/ecodeclub/ekit/retry"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitGormDB() *gorm.DB {
	url := "root:root@tcp(127.0.0.1:13306)/w-book-article?charset=utf8&parseTime=true&loc=Local"
	WaitForDBSetup(url)
	gormDb, err := gorm.Open(mysql.Open(url))
	if err != nil {
		panic(err)
	}

	db.InitTables(gormDb)
	return gormDb
}

func WaitForDBSetup(dsn string) {
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	const maxInterval = 10 * time.Second
	const maxRetries = 10
	strategy, err := retry.NewExponentialBackoffRetryStrategy(time.Second, maxInterval, maxRetries)
	if err != nil {
		panic(err)
	}

	const timeout = 5 * time.Second
	for {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		err = sqlDB.PingContext(ctx)
		cancel()
		if err == nil {
			break
		}
		next, ok := strategy.Next()
		if !ok {
			panic("WaitForDBSetup 重试失败......")
		}
		time.Sleep(next)
	}
}
