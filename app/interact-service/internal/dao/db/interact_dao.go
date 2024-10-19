package db

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Interaction struct {
	Id         int64  `json:"",gorm:"primaryKey,autoIncrement"`
	Biz        string `json:"type:varchar(128); uniqueIndex:biz_type_id"`
	BizId      int64  `json:"",gorm:"uniqueIndex:biz_type_id"`
	LikeCnt    int64  `json:""`
	ReadCnt    int64  `json:""`
	CollectCnt int64  `json:""`

	Ctime int64 `json:""`
	Utime int64 `json:""`
}

type InteractDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewInteractDao(db *gorm.DB) *InteractDao {
	return &InteractDao{db: db}
}

func (d *InteractDao) FindInteractByBiz(ctx context.Context, biz string, bizId int64) (*Interaction, error) {
	model := &Interaction{}

	result := d.db.Where("biz=? and biz_id=?", biz, bizId).First(model)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "查询Interaction失败")
	}
	return model, nil
}

func (d *InteractDao) IncreLike(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	interaction := &Interaction{
		Biz:     biz,
		BizId:   bizId,
		LikeCnt: 1,
		Utime:   now,
		Ctime:   now,
	}
	result := d.db.Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("`like_cnt`+1"),
				"utime":    now,
			}), // 更新字段
		},
	).Create(interaction)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *InteractDao) DecreLike(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()

	result := d.db.Where("biz=? and biz_id=?", biz, bizId).Updates(map[string]any{
		"like_cnt": gorm.Expr("`like_cnt` - 1"),
		"utime":    now,
	})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *InteractDao) IncreCollectCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	interaction := &Interaction{
		Biz:        biz,
		BizId:      bizId,
		CollectCnt: 1,
		Utime:      now,
		Ctime:      now,
	}
	result := d.db.Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"collect_cnt": gorm.Expr("`collect_cnt`+1"),
				"utime":       now,
			}), // 更新字段
		},
	).Create(interaction)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *InteractDao) DecreCollectCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()

	result := d.db.Where("biz=? and biz_id=?", biz, bizId).Updates(map[string]any{
		"collect_cnt": gorm.Expr("`collect_cnt` - 1"),
		"utime":       now,
	})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *InteractDao) IncreRead(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	interaction := &Interaction{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Utime:   now,
		Ctime:   now,
	}
	result := d.db.Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"read_cnt": gorm.Expr("`read_cnt`+1"),
				"utime":    now,
			}), // 更新字段
		},
	).Create(interaction)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *InteractDao) DecreReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()

	result := d.db.Where("biz=? and biz_id=?", biz, bizId).Updates(map[string]any{
		"read_cnt": gorm.Expr("`read_cnt` - 1"),
		"utime":    now,
	})

	if result.Error != nil {
		return result.Error
	}

	return nil
}
