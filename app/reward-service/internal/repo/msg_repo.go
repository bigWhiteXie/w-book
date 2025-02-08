package repo

import (
	"context"

	"codexie.com/w-book-common/repo"
	"codexie.com/w-book-reward/internal/dao/db"
	"gorm.io/gorm"
)

type IMessageRepository interface {
	// CreateMessage 创建消息记录
	CreateMessage(ctx context.Context, outTradeNo, status string) error
	// IsMessageProcessed 检查消息是否已处理
	IsMessageProcessed(ctx context.Context, outTradeNo, status string) (bool, error)
}

type MessageRepository struct {
	*repo.BaseRepo
}

func NewMessageRepository(gormdb *gorm.DB) *MessageRepository {
	// 自动迁移消息表结构
	if err := gormdb.AutoMigrate(&db.Message{}); err != nil {
		panic(err)
	}
	return &MessageRepository{
		repo.NewBaseRepo(gormdb),
	}
}

func (r *MessageRepository) CreateMessage(ctx context.Context, outTradeNo, status string) error {
	dao := db.NewMessageDao(ctx, r.GetDB())
	err := dao.CreateMessage(outTradeNo, status)
	if err == db.ErrMessageExists {
		return ErrMessageExists
	}
	return err
}

func (r *MessageRepository) IsMessageProcessed(ctx context.Context, outTradeNo, status string) (bool, error) {
	dao := db.NewMessageDao(ctx, r.GetDB())
	return dao.IsMessageProcessed(outTradeNo, status)
}

// 定义业务错误
var (
	ErrMessageExists = db.ErrMessageExists
)
