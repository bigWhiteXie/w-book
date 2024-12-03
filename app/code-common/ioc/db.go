package ioc

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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

func InitGormDB(mysqlConf MySQLConf) *gorm.DB {
	datasource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		mysqlConf.User,
		mysqlConf.Password,
		mysqlConf.Host,
		mysqlConf.Port,
		mysqlConf.Database,
		mysqlConf.CharSet,
		mysqlConf.ParseTime,
		mysqlConf.TimeZone)

	db, err := gorm.Open(mysql.Open(datasource), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   mysqlConf.Gorm.TablePrefix,   // such as: prefix_tableName
			SingularTable: mysqlConf.Gorm.SingularTable, // such as zero_user, not zero_users
		},
	})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(mysqlConf.Gorm.MaxOpenConns)
	sqlDB.SetMaxIdleConns(mysqlConf.Gorm.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(mysqlConf.Gorm.ConnMaxLifetime) * time.Second)

	return db
}
