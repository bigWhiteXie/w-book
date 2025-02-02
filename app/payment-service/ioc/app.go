package ioc

import (
	"time"

	"codexie.com/w-book-common/job"
	middleware "codexie.com/w-book-common/middleware/auth"
	"codexie.com/w-book-payment/internal/config"
	"codexie.com/w-book-payment/internal/handler"
	payJob "codexie.com/w-book-payment/internal/job"
	"codexie.com/w-book-payment/internal/repo"
	"codexie.com/w-book-payment/internal/server"
	"codexie.com/w-book-payment/internal/svc"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/rest"
)

type App struct {
	Server    *rest.Server
	RpcServer *server.PaymentServer
	JobCron   *job.JobCron
	SvcCtx    *svc.ServiceContext
}

func (app *App) Start() error {
	app.Server.Start()
	return nil
}

func (app *App) Stop() error {
	app.Server.Stop()
	return nil
}

func InitServer(c config.Config, payHandler *handler.AliPayHandler, redisClient *redis.Client) *rest.Server {
	// metric.InitMessageMetric(c.MetricConf)
	// logx.Infof("读取指标配置:%v", c.MetricConf)
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, payHandler)

	return server
}

func InitRpcServer(svcCtx *svc.ServiceContext) *server.PaymentServer {
	return server.NewPaymentServer(svcCtx)
}

func InitJobCron(payStatusJob *payJob.PayStatusJob, redisClient *redis.Client) *job.JobCron {
	cron := cron.New()
	jb := job.NewJobCron(cron, redisClient, "payment", 60*time.Second)
	jb.AddJob(payStatusJob, false)
	return jb
}

func InitPayStatusJob(svcCtx *svc.ServiceContext, repo *repo.PaymentRepository) *payJob.PayStatusJob {
	return payJob.NewPayStatusJob(svcCtx, repo)
}
