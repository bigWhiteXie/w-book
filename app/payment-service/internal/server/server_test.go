package server

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"testing"
	"time"

	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/kafka/producer"
	middleware "codexie.com/w-book-common/middleware/auth"
	"codexie.com/w-book-payment/api/pb"
	"codexie.com/w-book-payment/internal/config"
	"codexie.com/w-book-payment/internal/dao/db"
	"codexie.com/w-book-payment/internal/handler"
	"codexie.com/w-book-payment/internal/logic"
	"codexie.com/w-book-payment/internal/repo"
	"codexie.com/w-book-payment/internal/svc"
	"codexie.com/w-book-payment/pkg/constant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

func TestArticleGormHandler(t *testing.T) {
	suite.Run(t, &PaymentGprcSuite{})
}

type PaymentGprcSuite struct {
	suite.Suite
	db     *gorm.DB
	cache  *redis.Client
	server *zrpc.RpcServer
	client pb.PaymentClient
	c      config.Config
}

// 拷贝wire的代码进来
func (s *PaymentGprcSuite) SetupSuite() {
	var configFile = flag.String("f", "/usr/local/go_project/w-book/app/payment-service/etc/payment.yaml", "the config file")
	var config2 config.Config
	conf.MustLoad(*configFile, &config2)

	logx.Disable() // 禁用默认的日志配置
	logx.MustSetup(logx.LogConf{
		Mode:  "console", // 输出到控制台
		Level: "debug",   // 设置日志级别为 debug
	})

	aliPayConfig := config.GetAliPayConfig(config2)
	mySQLConf := config.GetMySQLConf(config2)
	db := ioc.InitGormDB(mySQLConf)
	paymentRepository := repo.NewPaymentRepository(db)
	payMsgRepo := repo.NewPayMsgRepo(db)
	kafkaConf := config.GetKafkaConf(config2)
	client := ioc.InitKafkaClient(kafkaConf)
	producerProducer := producer.NewKafkaProducer(client)
	paymentLogic := logic.NewPaymentLogic(paymentRepository, payMsgRepo, producerProducer)

	aliPayLogic := logic.NewAliPayLogic(aliPayConfig, paymentLogic)
	serviceContext := svc.NewServiceContext(config2, aliPayLogic)
	redisConf := config.GetRedisConf(config2)
	redisClient := ioc.InitRedis(redisConf)
	paymentServer := InitRpcServer(serviceContext)

	s.db = db
	s.cache = redisClient
	s.c = config2

	//启动grpc服务
	grpcServer := zrpc.MustNewServer(s.c.Grpc, func(grpcServer *grpc.Server) {
		pb.RegisterPaymentServer(grpcServer, paymentServer)
		if s.c.Mode == service.DevMode || s.c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.server = grpcServer

	go func() {
		fmt.Printf("Starting rpc server at %s...\n", s.c.Grpc.ListenOn)
		grpcServer.Start()
	}()
	time.Sleep(2 * time.Second)

	clientConf := zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: []string{"127.0.0.1:2479"},
			Key:   "payment.rpc",
		},
	}
	s.client = pb.NewPaymentClient(zrpc.MustNewClient(clientConf).Conn())
}

func (s *PaymentGprcSuite) TearDownTest() {
	s.db.Exec("truncate table `payment_record`")

	keys, _ := s.cache.Keys(context.Background(), "*").Result()
	for _, k := range keys {
		s.cache.Del(context.Background(), k)
	}

	s.server.Stop()
}

func (s *PaymentGprcSuite) TestZfbPrepay() {
	t := s.T()
	testCases := []testCase[int, *pb.PrepayReq]{
		{
			name: "支付宝预支付订单——文章打赏",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				dao := db.NewPaymentDao(context.Background(), s.db)
				record, err := dao.FindPaymentByBiz("reward-article", "1")
				assert.NoError(t, err)
				logx.Infow("payment record",
					logx.LogField{Key: "id", Value: record.Id},
					logx.LogField{Key: "Biz", Value: record.Biz},
					logx.LogField{Key: "OutTradeNo", Value: record.OutTradeNo},
					logx.LogField{Key: "Amt", Value: record.Amt},
					logx.LogField{Key: "platform", Value: record.Platform},
				)
			},
			req: &pb.PrepayReq{
				Biz:         "reward-article",
				OutTradeNo:  "1",
				Subject:     "文章打赏",
				TotalAmount: "888",
				Platform:    constant.ZFB,
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			res, err := s.client.PrePay(context.Background(), tc.req)
			if err != nil {
				panic(err)
			}
			//打印输出结果
			logx.Infow("invoke prepay",
				logx.LogField{Key: "payUrl", Value: res.PayUrl},
				logx.LogField{Key: "platform", Value: "zfb"},
			)
		})
	}
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,optional"`
}
type testCase[T any, R any] struct {
	name     string
	before   func(t *testing.T)
	after    func(t *testing.T)
	req      R
	wantCode int
	wantRes  *pb.PrepayResp
}

type App struct {
	Server    *rest.Server
	RpcServer *PaymentServer
	svc       *svc.ServiceContext
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
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, payHandler)

	return server
}

func InitRpcServer(svc *svc.ServiceContext) *PaymentServer {
	return NewPaymentServer(svc)
}
