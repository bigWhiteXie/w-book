package ioc

import (
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/handler"
	"codexie.com/w-book-user/internal/logic"
	"codexie.com/w-book-user/internal/repo/cache"
	"codexie.com/w-book-user/internal/repo/dao"
	"codexie.com/w-book-user/internal/svc"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/rest"
)

// Injectors from wire.go:

func NewApp(c config.Config, mysqlConf ioc.MySQLConf, redisConf ioc.RedisConf) (*rest.Server, error) {
	db := ioc.InitGormDB(mysqlConf)
	userDao := dao.NewUserDao(db)
	client := ioc.InitRedis(redisConf)
	userCache := cache.NewRedisUserCache(client)
	iUserRepository := InitUserRepo(userDao, userCache)
	codeClient := svc.CreateCodeRpcClient(c)
	userLogic := logic.NewUserLogic(c, iUserRepository, codeClient)
	userHandler := handler.NewUserHandler(userLogic)
	server := NewServer(c, userHandler, client)
	return server, nil
}

// wire.go:

var ServerSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewUserHandler)

var LogicSet = wire.NewSet(logic.NewUserLogic)

var CtxSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(InitUserRepo)

var DBSet = wire.NewSet(dao.NewUserDao, cache.NewRedisUserCache)

var BaseSet = wire.NewSet(ioc.InitGormDB, ioc.InitRedis, svc.CreateCodeRpcClient)
