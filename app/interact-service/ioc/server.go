package ioc

import (
	"codexie.com/w-book-common/kafka/consumer"
	"codexie.com/w-book-common/metric"
	"codexie.com/w-book-common/middleware"
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/event"
	"codexie.com/w-book-interact/internal/handler"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/server"
	"codexie.com/w-book-interact/internal/svc"
	"codexie.com/w-book-interact/internal/worker"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func NewServer(c config.Config, articleHandler *handler.InteractHandler, redisClient *redis.Client, readListener *consumer.BatchConsumer[domain.ReadEvent], createListener *event.CreateEventListener, topLikeWorker worker.Worker) *rest.Server {
	metric.InitMessageMetric(c.MetricConf)
	logx.Infof("读取指标配置:%v", c.MetricConf)
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, articleHandler)
	readListener.StartListner()
	createListener.StartListner()
	workerManager := worker.Manager{}
	workerManager.AddWorker(topLikeWorker)
	// workerManager.Start()
	return server
}

func NewRpcServer(serviceContext *svc.ServiceContext, interactLogic *logic.InteractLogic) *server.InteractionServer {
	return server.NewInteractionServer(serviceContext, interactLogic)
}
