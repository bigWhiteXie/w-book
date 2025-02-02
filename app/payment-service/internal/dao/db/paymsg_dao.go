package db

import (
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type PayMsg struct {
	Id         int64  `json:"",gorm:"primaryKey"`
	Topic      string `json:""`
	Msg        string `json:""`
	RetryTimes int    `json:""`
	Ctime      int64  `json:"",`
	Utime      int64  `json:"",gorm:"index:idx_utime"`
}

func (a *PayMsg) MarshalBinary() ([]byte, error) {
	return json.Marshal(a)
}

func (a *PayMsg) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

type PayMsgDao struct {
	db *gorm.DB
	logx.Logger
}

func NewPayMsgDao(ctx context.Context, db *gorm.DB) *PayMsgDao {
	return &PayMsgDao{
		db:     db,
		Logger: logx.WithContext(ctx),
	}
}

func (dao *PayMsgDao) FindByLastId(lastId int64, limit int) ([]PayMsg, error) {
	var msgs []PayMsg
	result := dao.db.Where("id > ?", lastId).
		Order("id ASC").
		Limit(limit).
		Find(&msgs)
	if result.Error != nil {
		return nil, result.Error
	}
	return msgs, nil
}

func (dao *PayMsgDao) DeleteByIds(ids []int64) error {
	result := dao.db.Where("id IN ?", ids).Delete(&PayMsg{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (dao *PayMsgDao) CreatePayMsg(topic, msg string) error {
	now := time.Now()
	payMsg := &PayMsg{
		Topic: topic,
		Msg:   msg,
		Ctime: now.UnixMilli(),
		Utime: now.UnixMilli(),
	}
	result := dao.db.Create(payMsg)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
