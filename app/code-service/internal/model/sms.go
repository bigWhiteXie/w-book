package model

import (
	"encoding/json"
	"time"
)

type SmsSendRecord struct {
	ID          int       `gorm:"primaryKey;autoIncrement"`
	Phone       string    `gorm:"size:20;not null"`
	Content     string    `gorm:"type:text;not null"` // 存储 JSON
	Status      int       `gorm:"not null"`           // 0 pending 1 processing 2 success 4 fail
	RetryCount  int       `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	ErrorMsg    string    `gorm:"type:text"`
	NextRetryAt time.Time `gorm:"index:idx_status_next_retry"` // 你忘记在表结构里添加这个字段，我补充了
}

func NewSmsRecord(phone string, content map[string]string) *SmsSendRecord {
	cnt, _ := json.Marshal(content)
	now := time.Now()
	return &SmsSendRecord{
		Phone:      phone,
		Content:    string(cnt),
		Status:     0,
		RetryCount: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
