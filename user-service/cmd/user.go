package main

import (
	"flag"
	"fmt"
	"net/http"

	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/handler"
	"codexie.com/w-book-user/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "/usr/local/go_project/w-book/user-service/etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCustomCors(nil,  rest.WithCors())
	server.Use(middleware.NewLimiterMiddleware(limiter.NewRateLimiter(c.IpRate)).Handle)

	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
