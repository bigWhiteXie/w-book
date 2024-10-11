package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/handler"
	"codexie.com/w-book-user/internal/logic"
	"codexie.com/w-book-user/internal/repo"
	"codexie.com/w-book-user/internal/repo/cache"
	"codexie.com/w-book-user/internal/repo/dao"
	"codexie.com/w-book-user/internal/svc"
	"codexie.com/w-book-user/pkg/ijwt"
	"codexie.com/w-book-user/pkg/limiter"
	"codexie.com/w-book-user/pkg/middleware"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "/usr/local/go_project/w-book/app/user-service/etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	server, err := NewApp(c)
	if err != nil {
		panic(err)
	}
	defer server.Stop()
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

func NewServer(c config.Config, userHandler *handler.UserHandler) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewLimiterMiddleware(limiter.NewRateLimiter(c.IpRate)).Handle)
	server.Use(middleware.NewJwtMiddleware().Handle)
	handler.RegisterHandlers(server, c, userHandler)
	return server
}

func NewApp(c config.Config) (*rest.Server, error) {
	db := svc.CreteDbClient(c)
	userDao := dao.NewUserDao(db)
	client := svc.CreateRedisClient(c)
	userCache := cache.NewRedisUserCache(client)
	ijwt.InitJwtHandler(client)
	iUserRepository := repo.NewUserRepository(userDao, userCache)
	codeClient := svc.CreateCodeRpcClient(c)
	iUserLogic := logic.NewUserLogic(c, iUserRepository, codeClient)
	userHandler := handler.NewUserHandler(iUserLogic)
	server := NewServer(c, userHandler)
	return server, nil
}
