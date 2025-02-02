package config

import (
	"codexie.com/w-book-common/ioc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Grpc         zrpc.RpcServerConf
	AliPayConfig AliPayConfig
	KafkaConf    ioc.KafkaConf
	MySQLConf    ioc.MySQLConf
	RedisConf    ioc.RedisConf
	// MetricConf metric.ConstMetricLabelsConf
}

func GetAliPayConfig(c Config) AliPayConfig {
	return c.AliPayConfig
}

func GetMySQLConf(c Config) ioc.MySQLConf {
	return c.MySQLConf
}

func GetKafkaConf(c Config) ioc.KafkaConf {
	return c.KafkaConf
}

// func GetMetricConf(c Config) metric.ConstMetricLabelsConf {
// 	return c.MetricConf
// }

func GetRedisConf(c Config) ioc.RedisConf {
	return c.RedisConf
}

type AliPayConfig struct {
	AppId        string `json:",optional"`
	PrivateKey   string `json:",optional"`
	IsProduction bool   `json:",optional"`
}
