package ioc

import (
	"codexie.com/w-book-common/limiter"
	"codexie.com/w-book-common/middleware"
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/handler"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
)

func NewServer(c config.Config, userHandler *handler.UserHandler, client *redis.Client) *rest.Server {
	bucketLimiter := limiter.NewTokenBucketLimiter(client, &c.BucketRateConf)
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	server.Use(middleware.NewLimiterMiddleware(limiter.NewRateLimiter(c.IpRate)).Handle)
	server.Use(middleware.NewBucketLimiterMiddleware(bucketLimiter).Handle)

	server.Use(middleware.NewJwtMiddleware().Handle)
	handler.RegisterHandlers(server, c, userHandler)

	return server
}
