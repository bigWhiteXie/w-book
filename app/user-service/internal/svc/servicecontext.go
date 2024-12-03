package svc

import (
	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/model"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config        config.Config
	CodeRpcClient pb.CodeClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:        c,
		CodeRpcClient: CreateCodeRpcClient(c),
	}
}

func CreateCodeRpcClient(c config.Config) pb.CodeClient {
	return pb.NewCodeClient(zrpc.MustNewClient(c.CodeRpcConf).Conn())
}

func InitTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return err
	}

	return nil
}
