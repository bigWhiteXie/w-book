package config

import (
	"codexie.com/w-book-common/ioc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	InteractRpcConf zrpc.RpcClientConf
	MySQLConf       ioc.MySQLConf
	KafkaConf       ioc.KafkaConf
	RedisConf       ioc.RedisConf
}
