package ioc

import (
	"codexie.com/w-book-common/kafka/consumer"
	"codexie.com/w-book-common/metric"
	"codexie.com/w-book-common/middleware"
	"codexie.com/w-book-interact/internal/config"
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

type App struct {
	Server        *rest.Server
	Consumers     []consumer.Consumer
	TopLikeWorker *worker.TopLikeWorker
}

func (app *App) Start() error {
	for _, consumer := range app.Consumers {
		consumer.Start()
	}
	app.Server.Start()
	return nil
}

func (app *App) Stop() error {
	for _, consumer := range app.Consumers {
		consumer.Stop()
	}
	app.Server.Stop()
	return nil
}


func InitServer(c config.Config, articleHandler *handler.InteractHandler, redisClient *redis.Client) *rest.Server {
	metric.InitMessageMetric(c.MetricConf)
	logx.Infof("读取指标配置:%v", c.MetricConf)
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, articleHandler)
	return server
}

func InitConsumers(readListener *event.ReadEvtListener, createListener *event.CreateEventListener) []consumer.Consumer {
	consumers := make([]consumer.Consumer, 0, 2)
	consumers = append(consumers, readListener)
	consumers = append(consumers, createListener)
	return consumers
}

func InitRpcServer(serviceContext *svc.ServiceContext, interactLogic *logic.InteractLogic) *server.InteractionServer {
	return server.NewInteractionServer(serviceContext, interactLogic)
}
