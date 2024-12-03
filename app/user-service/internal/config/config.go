package config

import (
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/limiter"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct { // JWT 认证需要的密钥和过期时间配置
		AccessSecret string
		AccessExpire int64
	}
	BucketRateConf limiter.TokenBucketRateConf
	CodeRpcConf    zrpc.RpcClientConf
	MySQLConf      ioc.MySQLConf
	RedisConf      ioc.RedisConf
	IpRate         limiter.IpLimitConfig
}
