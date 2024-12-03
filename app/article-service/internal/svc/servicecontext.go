package svc

import (
	"codexie.com/w-book-article/internal/config"
	dao "codexie.com/w-book-article/internal/dao/db"

	"codexie.com/w-book-interact/api/pb/interact"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Cache  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {

	return &ServiceContext{
		Config: c,
	}
}

func CreateCodeRpcClient(c config.Config) interact.InteractionClient {
	return interact.NewInteractionClient(zrpc.MustNewClient(c.InteractRpcConf).Conn())
}

func InitTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&dao.Article{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&dao.PublishedArticle{}); err != nil {
		return err
	}
	return nil
}
