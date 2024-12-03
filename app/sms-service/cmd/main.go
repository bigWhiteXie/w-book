package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-code/ioc"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-code/internal/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/prometheus"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "/usr/local/go_project/w-book/app/sms-service/etc/sms.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	app, err := ioc.NewApp(c, c.MySQLConf, c.RedisConf, c.KafkaConf, c.SmsConf, c.MetricConf)
	if err != nil {
		panic(err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterSMSServer(grpcServer, app.Server)
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	defer func() {
		app.SmsEvtListener.Stop()
		s.Stop()
	}()
	go func() {
		app.SmsEvtListener.StartConsumer()
	}()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	// 启动prometheus
	go func() {
		prometheus.StartAgent(c.ServiceConf.Prometheus)
	}()
	s.Start()
}
