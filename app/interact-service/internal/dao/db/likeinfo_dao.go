package db

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ArticleStatusPublished = 1
)

type LikeInfo struct {
	Id     int64  `json:"",gorm:"primaryKey"`
	Biz    string `json:"type:varchar(128); uniqueIndex:biz_type_id"`
	BizId  int64  `json:"",gorm:"uniqueIndex:biz_type_id"`
	Uid    int64  `json:"",gorm:""`
	Status uint8  `json:"",gorm:""`
	Ctime  int64  `json:"",`
	Utime  int64  `json:""`
}

func (a *LikeInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(a)
}

func (a *LikeInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

type LikeInfoDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewLikeInfoDao(db *gorm.DB) *LikeInfoDao {
	return &LikeInfoDao{db: db}
}

func (d *LikeInfoDao) UpdateLikeInfo(ctx context.Context, uid int64, biz string, bizId int64, status uint8) error {
	now := time.Now().UnixMilli()
	likeInfo := &LikeInfo{
		Biz:    biz,
		BizId:  bizId,
		Uid:    uid,
		Status: uint8(status),
		Ctime:  now,
		Utime:  now,
	}
	// 尝试插入点赞记录，若冲突则更新status和utime
	res := d.db.Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"status": status,
				"utime":  now,
			}), // 更新字段
		},
	).Create(likeInfo)

	if res.Error != nil {
		return res.Error
	}
	return nil
}
