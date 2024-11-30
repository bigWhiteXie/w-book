package db

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Record struct {
	Id    int64  `json:"" gorm:"primaryKey;autoIncrement"`
	Biz   string `json:"" gorm:"type:varchar(128);index:idx_biz_uid_utime;uniqueIndex:uk_biz_bizid_uid"` // 在唯一索引和查询索引中使用
	BizId int64  `json:"" gorm:"uniqueIndex:uk_biz_bizid_uid;index:idx_biz_uid_utime"`
	UId   int64  `json:"" gorm:"uniqueIndex:uk_biz_bizid_uid;index:idx_uid_utime;index:idx_biz_uid_utime"`

	Ctime int64 `json:""`
	Utime int64 `json:"" gorm:"index:idx_uid_utime,sort:desc;index:idx_biz_uid_utime,sort:desc"` // 为两个不同的查询创建排序索引
}

type RecordDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewRecordDao(db *gorm.DB) *RecordDao {
	return &RecordDao{db: db}
}

func (d *RecordDao) AddRecords(ctx context.Context, bizs []string, bizIds []int64, uids []int64) error {
	now := time.Now().UnixMilli()
	return d.db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(bizs); i++ {
			biz := bizs[i]
			bizId := bizIds[i]
			uid := uids[i]
			record := &Record{
				Biz:   biz,
				BizId: bizId,
				UId:   uid,
				Utime: now,
				Ctime: now,
			}
			result := tx.Clauses(
				clause.OnConflict{
					DoUpdates: clause.Assignments(map[string]any{
						"utime": now,
					}), // 更新字段
				},
			).Create(record)
			if result.Error != nil {
				return errors.Wrapf(result.Error, "[RecordDao_AddRecords] 创建资源浏览记录失败,biz:%s,bizId:%d,uid:%d", biz, bizId, uid)
			}
		}
		return nil
	})
}

func (d *RecordDao) ListRecordByBiz(ctx context.Context, biz string, bizId int64, limit, offset int) ([]*Record, error) {
	var records []*Record
	err := d.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ?", biz, bizId).
		Order("utime DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error

	if err != nil {
		return nil, errors.Wrapf(err, "[RecordDao_ListRecordByBiz] 查询资源浏览记录失败,biz:%s,bizId:%d", biz, bizId)
	}
	return records, nil
}

func (d *RecordDao) ListRecordByUid(ctx context.Context, uid int64, limit, offset int) ([]*Record, error) {
	var records []*Record
	err := d.db.WithContext(ctx).
		Where("uid = ?", uid).
		Order("utime DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error

	if err != nil {
		return nil, errors.Wrapf(err, "[RecordDao_ListRecordByUid] 查询资源浏览记录失败,uid:%d", uid)
	}
	return records, nil
}
