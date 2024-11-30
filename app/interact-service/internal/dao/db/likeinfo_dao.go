package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ArticleStatusPublished = 1
)

type LikeInfo struct {
	Id     int64  `json:"" gorm:"primaryKey"`
	Biz    string `json:"" gorm:"type:varchar(122); uniqueIndex:biz_uid_idx"`
	BizId  int64  `json:"" gorm:"uniqueIndex:biz_uid_idx"`
	Uid    int64  `json:"" gorm:""`
	Status uint8  `json:"" gorm:""`
	Ctime  int64  `json:""`
	Utime  int64  `json:"" gorm:"uniqueIndex:biz_uid_idx`
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
		return errors.Wrapf(res.Error, "[LikeInfoDao_UpdateLikeInfo] 插入点赞信息失败")
	}
	return nil
}

func (d *LikeInfoDao) FindLikeInfo(ctx context.Context, uid int64, biz string, bizId int64) (*LikeInfo, error) {
	likeInfo := &LikeInfo{}
	res := d.db.Where("biz=? and biz_id=? and uid=?", biz, bizId, uid).First(likeInfo)
	if res.Error != nil {
		return nil, errors.Wrapf(res.Error, "[LikeInfoDao_FindLikeInfo] 查询点赞信息失败")
	}

	return likeInfo, nil
}
