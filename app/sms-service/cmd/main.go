package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-code/internal/kafka/consumer"
	"codexie.com/w-book-code/internal/logic"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-code/internal/repo/cache"
	"codexie.com/w-book-code/internal/repo/dao"
	"codexie.com/w-book-common/metric"
	"codexie.com/w-book-common/producer"

	"codexie.com/w-book-code/pkg/sms"
	"codexie.com/w-book-code/pkg/sms/provider"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-code/internal/config"
	"codexie.com/w-book-code/internal/server"
	"codexie.com/w-book-code/internal/svc"

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
	ctx := svc.NewServiceContext(c)
	metric.InitMessageMetric(c.MetricConf)
	smsRepo := repo.NewSmsRepo(cache.NewCodeRedisCache(ctx.Cache), dao.NewCodeDao(ctx.DB))
	producer := producer.NewKafkaProducer(ctx.KafkaProvider)
	prometheusSmsService := NewSmsSerice(c.SmsConf, c.MetricConf)
	asyncSmsLogic := logic.NewASyncSmsLogic(prometheusSmsService, smsRepo, producer)
	smsConsumer := consumer.NewSmsConsumer(c.KafkaConf.Topic, ctx.ConsumerGroup, smsRepo, prometheusSmsService)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterSMSServer(grpcServer, server.NewSMSServer(ctx, smsRepo, producer, asyncSmsLogic))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer smsConsumer.Stop()
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	go func() {
		smsConsumer.StartConsumer()
	}()

	// 启动prometheus
	go func() {
		prometheus.StartAgent(c.ServiceConf.Prometheus)
	}()
	s.Start()
}

func NewSmsSerice(conf sms.SmsConf, labelConf metric.ConstMetricLabelsConf) logic.SmsService {
	mem := provider.NewMemoryClient(conf.Memory)
	tc := provider.NewTCSmsClient(conf.TC)
	providerSmsLogic := logic.NewProviderSmsLogic(mem, tc)
	return logic.NewPrometheusSmsLogic(labelConf, providerSmsLogic)
}
