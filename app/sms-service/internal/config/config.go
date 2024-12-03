package config

import (
	"codexie.com/w-book-code/pkg/sms"
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/metric"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MetricConf metric.ConstMetricLabelsConf
	SmsConf    sms.SmsConf `json:",optional"`
	MySQLConf  ioc.MySQLConf
	RedisConf  ioc.RedisConf
	KafkaConf  ioc.KafkaConf
}
