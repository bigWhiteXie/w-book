package repo

import (
	"context"

	"codexie.com/w-book-common/repo"
	"codexie.com/w-book-payment/internal/dao/db"
	"gorm.io/gorm"
)

type IPayMsgRepo interface {
	CreateMsg(ctx context.Context, topic string, msg string) error
}

type PayMsgRepo struct {
	*repo.BaseRepo
}

func NewPayMsgRepo(gormDB *gorm.DB) *PayMsgRepo {
	if err := gormDB.AutoMigrate(&db.PayMsg{}); err != nil {
		panic(err)
	}
	return &PayMsgRepo{
		repo.NewBaseRepo(gormDB),
	}
}

func (repo *PayMsgRepo) CreateMsg(ctx context.Context, topic string, msg string) error {
	msgDao := db.NewPayMsgDao(ctx, repo.GetDB())
	return msgDao.CreatePayMsg(topic, msg)
}
