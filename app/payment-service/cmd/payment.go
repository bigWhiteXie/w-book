package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-payment/api/pb"
	"codexie.com/w-book-payment/internal/config"
	"codexie.com/w-book-payment/ioc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "/usr/local/go_project/w-book/app/payment-service/etc/payment.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	app, err := ioc.NewPaymentApp(c)
	if err != nil {
		panic(err)
	}
	// 启动jobCron
	app.JobCron.Start()
	//启动rpc服务
	go func() {
		s := zrpc.MustNewServer(c.Grpc, func(grpcServer *grpc.Server) {
			pb.RegisterPaymentServer(grpcServer, app.RpcServer)
			if c.Mode == service.DevMode || c.Mode == service.TestMode {
				reflection.Register(grpcServer)
			}
		})
		defer s.Stop()

		fmt.Printf("Starting rpc server at %s...\n", c.Grpc.ListenOn)
		s.Start()
	}()

	//启动httpserver、消费者
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	app.Start()
	defer app.Stop()
}
