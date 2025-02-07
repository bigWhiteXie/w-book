package config

import (
	"codexie.com/w-book-common/ioc"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQLConf ioc.MySQLConf
}

func GetMySQLConf(c Config) ioc.MySQLConf {
	return c.MySQLConf
}
