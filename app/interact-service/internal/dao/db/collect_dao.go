package db

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
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
	Id    int64  `json:"" gorm:"primaryKey"`
	Uid   int64  `json:""`
	Biz   string `json:"" gorm:"uniqueIndex:biz_cid_idx"`
	BizId int64  `json:"" gorm:"uniqueIndex:biz_cid_idx"`
	Cid   int64  `json:"" gorm:"index:cid_ctime_idx;uniqueIndex:biz_cid_idx"`
	Name  string `json:""`
	Ctime int64  `json:"" gorm:"index:cid_ctime_idx"`
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
		//设置唯一索引，重复点赞会异常
		return tx.Create(entity).Error
	})
	if err != nil {
		return nil, errors.Wrap(err, "[CollectionDao] AddCollectionItem")
	}
	return entity, err
}

func (dao *CollectionDao) DelCollectionItem(ctx context.Context, entity *CollectionItem) (*CollectionItem, error) {
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Collection{}).Where("uid=? and id=? and count >= 1", entity.Uid, entity.Cid).Updates(map[string]any{
			"count": gorm.Expr("`count`- 1"),
			"utime": time.Now().UnixMilli(),
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("用户[%d]往他人收藏夹[%d]中删除数据", entity.Uid, entity.Cid)
		}
		delRes := tx.Where("id=?", entity.Id).Delete(&CollectionItem{})
		if err := delRes.Error; err != nil {
			return err
		}

		if delRes.RowsAffected == 0 {
			return errors.Errorf("用户[%d]重复删除收藏数据[%d]", entity.Uid, entity.Id)
		}
		return nil
	})

	return entity, err
}

func (dao *CollectionDao) FindCollection(ctx context.Context, uid, bizId int64, biz string) (*CollectionItem, error) {
	item := &CollectionItem{}
	res := dao.db.Where("biz=? and biz_id=? and uid=?", biz, bizId, uid).First(item)
	return item, res.Error
}
