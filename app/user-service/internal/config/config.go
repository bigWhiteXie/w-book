package config

import (
	"codexie.com/w-book-user/pkg/limiter"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct { // JWT 认证需要的密钥和过期时间配置
		AccessSecret string
		AccessExpire int64
	}
	CodeRpcConf zrpc.RpcClientConf
	MySQLConf   MySQLConf
	RedisConf   RedisConf
	IpRate      limiter.IpLimitConfig
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

	//DefaultStringSize         uint          `json:"" yaml:"DefaultStringSize"`         // string 类型字段的默认长度
	AutoMigrate bool `json:"" yaml:"AutoMigrate"`
	//DisableDatetimePrecision  bool          `json:"" yaml:"DisableDatetimePrecision"`  // 禁用 datetime 精度
	//SkipInitializeWithVersion bool          `json:"" yaml:"SkipInitializeWithVersion"` // 根据当前 MySQL 版本自动配置
	//
	//SlowSql                   time.Duration `json:"" yaml:"SlowSql"`                   //慢SQL
	//LogLevel                  string        `json:"" yaml:"LogLevel"`                  // 日志记录级别
	//IgnoreRecordNotFoundError bool          `json:"" yaml:"IgnoreRecordNotFoundError"` // 是否忽略ErrRecordNotFound(未查到记录错误)

	Gorm GormConf `json:"" yaml:"Gorm"`
}

type RedisConf struct {
	Host string `json:"host"`
	Type string `json:",default=node,options=node|cluster"`
	Pass string `json:",optional"`
	Tls  bool   `json:",optional"`
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
