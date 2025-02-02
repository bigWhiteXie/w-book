package ioc

import (
	"codexie.com/w-book-payment/api/pb"
	"codexie.com/w-book-reward/internal/config"
	"codexie.com/w-book-reward/internal/handler"

	middleware "codexie.com/w-book-common/middleware/auth"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type App struct {
	Server *rest.Server
}

func InitServer(c config.Config, rewardHandler *handler.RewardHandler, redisClient *redis.Client) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, rewardHandler)
	return server
}

func InitPaymentRpcClient(rpcConf zrpc.RpcClientConf) pb.PaymentClient {
	return pb.NewPaymentClient(zrpc.MustNewClient(rpcConf).Conn())
}
