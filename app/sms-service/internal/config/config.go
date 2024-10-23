package config

import (
	"codexie.com/w-book-code/pkg/sms"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	SmsConf   sms.SmsConf `json:",optional"`
	MySQLConf MySQLConf
	RedisConf RedisConf
	KafkaConf KafkaConf
}

type MySQLConf struct {
	Host      string `json:"" yaml:"Host"`
	Port      int64  `json:"" yaml:"Port"`
	User      string `json:"" yaml:"User"`
	Password  string `json:"" yaml:"Password"`
	Database  string `json:"" yaml:"Database"`
	CharSet   string `json:"" yaml:"CharSet"`
	TimeZone  string `json:"" yaml:"TimeZone"`
	ParseTime bool   `json:"" yaml:"ParseTime"`
	Enable    bool   `json:"" yaml:"Enable"` // use mysql or not

	AutoMigrate bool `json:"" yaml:"AutoMigrate"`

	Gorm GormConf `json:"" yaml:"Gorm"`
}

type RedisConf struct {
	Host string `json:"host"`
	Type string `json:",default=node,options=node|cluster"`
	Pass string `json:",optional"`
	Tls  bool   `json:",optional"`
}

type KafkaConf struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}

// gorm config
type GormConf struct {
	//SkipDefaultTx   bool   `json:"" yaml:"SkipDefaultTx"`                            //是否跳过默认事务
	//CoverLogger     bool   `json:"" yaml:"CoverLogger"`                              //是否覆盖默认logger
	//PreparedStmt    bool   `json:"" yaml:"PreparedStmt"`                              // 设置SQL缓存
	//CloseForeignKey bool   `json:"" yaml:"CloseForeignKey"` 						// 禁用外键约束
	SingularTable   bool   `json:"" yaml:"SingularTable"`        //是否使用单数表名(默认复数)，启用后，User结构体表将是user
	TablePrefix     string `json:",optional" yaml:"TablePrefix"` // 表前缀
	MaxOpenConns    int    `json:",default=1000" yaml:"MaxOpenConns"`
	MaxIdleConns    int    `json:",default=100" yaml:"MaxIdleConns"`
	ConnMaxLifetime int    `json:"" yaml:"ConnMaxLifetime"`
}
