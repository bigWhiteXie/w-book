package config

import (
	"codexie.com/w-book-common/ioc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type AccountRpcConf zrpc.RpcClientConf
type PaymentRpcConf zrpc.RpcClientConf

type Config struct {
	rest.RestConf

	PaymentRpcConf PaymentRpcConf
	AccountRpcConf AccountRpcConf
	MySQLConf      ioc.MySQLConf
	KafkaConf      ioc.KafkaConf
	RedisConf      ioc.RedisConf
}

func GetPaymentRpcConf(c Config) PaymentRpcConf {
	return c.PaymentRpcConf
}

func GetAccountRpcConf(c Config) AccountRpcConf {
	return c.AccountRpcConf
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
