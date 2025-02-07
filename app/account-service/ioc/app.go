package ioc

import (
	"codexie.com/w-book-account/internal/dao/db"
	"codexie.com/w-book-account/internal/server"
	"codexie.com/w-book-account/internal/service"
	"codexie.com/w-book-account/internal/svc"
	"codexie.com/w-book-common/ioc"
	"gorm.io/gorm"
)

type App struct {
	RpcServer *server.AccountServer
	SvcCtx    *svc.ServiceContext
}

func (app *App) Start() error {

	return nil
}

func (app *App) Stop() error {
	return nil
}

func InitRpcServer(svcCtx *svc.ServiceContext, accountService *service.AccountService) *server.AccountServer {
	return server.NewAccountServer(accountService)
}

func InitDB(conf ioc.MySQLConf) *gorm.DB {
	gormDB := ioc.InitGormDB(conf)
	if err := gormDB.AutoMigrate(&db.Account{}, &db.AccountActivity{}); err != nil {
		panic(err)
	}
	return gormDB
}
