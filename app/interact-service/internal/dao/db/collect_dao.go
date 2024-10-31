package db

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type Collection struct {
	Id    int64  `json:"",gorm:"primaryKey"`
	Name  string `json:"",gorm:"type:varchar(128)`
	Uid   int64  `json:"",gorm:"index:uid_ctime_idx"`
	Count int64  `json:"",gorm:""`
	Ctime int64  `json:"",gorm:"index:uid_ctime_idx",`
	Utime int64  `json:""`
}

type CollectionItem struct {
	Id    int64  `json:"",gorm:"primaryKey"`
	Uid   int64  `json:""`
	Cid   int64  `json:"",gorm:"index:cid_ctime_idx"`
	Biz   string `json:""`
	BizId int64  `json:""`
	Name  string `json:""`
	Ctime int64  `json:"",gorm:"index:cid_ctime_idx"`
	Utime int64  `json:""`
}

type CollectionDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewCollectionDao(db *gorm.DB) *CollectionDao {
	return &CollectionDao{db: db}
}

func (dao *CollectionDao) AddCollection(ctx context.Context, entity *Collection) error {
	entity.Id = 0
	return dao.db.Create(entity).Error
}

func (dao *CollectionDao) DelCollection(ctx context.Context, uid, cid int64) error {
	res := dao.db.Where("id=? and uid=?", cid, uid).Delete(&Collection{})
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		logx.Errorf("用户[%d]删除不属于他的收藏夹[%d]", uid, cid)
	}

	return nil
}

func (dao *CollectionDao) AddCollectionItem(ctx context.Context, entity *CollectionItem) (*CollectionItem, error) {
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Collection{}).Where("uid=? and id=?", entity.Uid, entity.Cid).Updates(map[string]any{
			"count": gorm.Expr("`count`+1"),
			"utime": time.Now().UnixMilli(),
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("用户[%d]往他人收藏夹[%d]中添加数据", entity.Uid, entity.Cid)
		}
		entity.Id = 0
		return tx.Create(entity).Error
	})

	return entity, err
}

func (dao *CollectionDao) DelCollectionItem(ctx context.Context, entity *CollectionItem) (*CollectionItem, error) {
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Collection{}).Where("uid=? and id=?", entity.Uid, entity.Cid).Updates(map[string]any{
			"count": gorm.Expr("`count`- 1"),
			"utime": time.Now().UnixMilli(),
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("用户[%d]往他人收藏夹[%d]中删除数据", entity.Uid, entity.Cid)
		}
		entity.Id = 0
		return tx.Where("id=?", entity.Id).Delete(&CollectionItem{}).Error
	})

	return entity, err
}
