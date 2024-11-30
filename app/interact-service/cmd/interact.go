package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-interact/api/pb/interact"
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/ioc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "/usr/local/go_project/w-book/app/interact-service/etc/interact-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server, err := ioc.NewApp(c)
	if err != nil {
		panic(err)
	}

	//启动rpc服务
	go func() {
		s := zrpc.MustNewServer(c.Grpc, func(grpcServer *grpc.Server) {
			server, _ := ioc.NewRpcApp(c)
			interact.RegisterInteractionServer(grpcServer, server)
			if c.Mode == service.DevMode || c.Mode == service.TestMode {
				reflection.Register(grpcServer)
			}
		})
		defer s.Stop()

		fmt.Printf("Starting rpc server at %s...\n", c.Grpc.ListenOn)
		s.Start()
	}()

	//启动api服务
	defer server.Stop()
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
