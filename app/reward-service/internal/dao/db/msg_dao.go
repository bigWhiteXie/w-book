package db

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Message 消息表
type Message struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`
	OutTradeNo string `gorm:"uniqueIndex:idx_tradeno_status"`
	Status     string `gorm:"uniqueIndex:idx_tradeno_status"`
	Ctime      int64
	Utime      int64
}

type MessageDao struct {
	db  *gorm.DB
	log logx.Logger
	ctx context.Context
}

func NewMessageDao(ctx context.Context, db *gorm.DB) *MessageDao {
	return &MessageDao{
		db:  db,
		log: logx.WithContext(ctx),
		ctx: ctx,
	}
}

// CreateMessage 创建消息记录，如果已存在则返回错误
func (dao *MessageDao) CreateMessage(outTradeNo, status string) error {
	now := time.Now().UnixMilli()
	message := &Message{
		OutTradeNo: outTradeNo,
		Status:     status,
		Ctime:      now,
		Utime:      now,
	}

	result := dao.db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(message)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// IsMessageProcessed 检查消息是否已经处理过
func (dao *MessageDao) IsMessageProcessed(outTradeNo, status string) (bool, error) {
	var count int64
	err := dao.db.Model(&Message{}).
		Where("out_trade_no = ? AND status = ?", outTradeNo, status).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

var (
	ErrMessageExists = gorm.ErrDuplicatedKey
)
