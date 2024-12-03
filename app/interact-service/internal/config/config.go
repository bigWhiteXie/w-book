package config

import (
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/metric"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Grpc       zrpc.RpcServerConf
	KafkaConf  ioc.KafkaConf
	MySQLConf  ioc.MySQLConf
	RedisConf  ioc.RedisConf
	MetricConf metric.ConstMetricLabelsConf
}
