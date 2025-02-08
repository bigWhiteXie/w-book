package ioc

import (
	"time"

	acocount_pb "codexie.com/w-book-account/api/pb"
	"codexie.com/w-book-payment/api/pb"

	"codexie.com/w-book-reward/internal/config"
	"codexie.com/w-book-reward/internal/event"
	"codexie.com/w-book-reward/internal/handler"

	middleware "codexie.com/w-book-common/middleware/auth"
	"codexie.com/w-book-common/middleware/filter"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type App struct {
	Server                 *rest.Server
	PayCallbackEvtListener *event.PayCallbackEvtListener
}

func InitServer(c config.Config, rewardHandler *handler.RewardHandler, redisClient *redis.Client) *rest.Server {
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewJwtMiddleware(redisClient).Handle)
	handler.RegisterHandlers(server, rewardHandler)
	return server
}

func InitPaymentRpcClient(rpcConf config.PaymentRpcConf) pb.PaymentClient {
	return pb.NewPaymentClient(zrpc.MustNewClient(zrpc.RpcClientConf(rpcConf)).Conn())
}

func InitAccountRpcClient(rpcConf config.AccountRpcConf) acocount_pb.AccountClient {
	return acocount_pb.NewAccountClient(zrpc.MustNewClient(zrpc.RpcClientConf(rpcConf)).Conn())
}

func InitRedisBloomFilter(redisClient *redis.Client) *filter.WindowedRedisBloomFilter {
	return filter.NewWindowedRedisBloomFilter(redisClient, "reward:bloom:window:", time.Hour, 10000, 0.01)
}
