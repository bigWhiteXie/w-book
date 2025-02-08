package db

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

var (
	WithoutIdErr      = errors.New("id must specified")
	NoRowsAffectedErr = errors.New("no rows affected")
	IdempotentErr     = errors.New("record is idempotent")
)

type RewardRecord struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`                                // 主键，自增
	Uid        int64  `gorm:"column:uid"`                                              // 打赏人ID
	AuthorId   int64  `gorm:"column:author_id"`                                        // 作者ID
	Amt        int64  `gorm:"column:amt"`                                              // 金额
	Biz        string `gorm:"column:biz;index:idx_biz_resource_id,priority:1"`         // 资源类型，联合索引
	ResourceId int64  `gorm:"column:resource_id;index:idx_biz_resource_id,priority:2"` // 资源ID，联合索引
	OutTradeNo string `gorm:"column:out_trade_no;uniqueIndex:idx_trade_no"`            // 外贸编号，唯一索引
	Platform   string `gorm:"column:platform"`                                         // 平台
	PayStatus  string `gorm:"column:pay_status"`                                       // 状态
	Ctime      int64  `gorm:"column:ctime"`                                            // 创建时间
	Utime      int64  `gorm:"column:utime;index:idx_uid_uptime"`                       // 更新时间，索引
}

type RewardDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db     *gorm.DB
	ctx    context.Context
	logger logx.Logger
}

func NewRewardDao(ctx context.Context, db *gorm.DB) *RewardDao {
	return &RewardDao{ctx: ctx, db: db, logger: logx.WithContext(ctx)}
}

func (dao *RewardDao) CreateReward(reward *RewardRecord) (int64, error) {
	result := dao.db.Create(reward)
	return reward.Id, result.Error
}

// UpdateRewardById 根据ID更新打赏记录
func (dao *RewardDao) UpdateRewardById(reward *RewardRecord) error {
	// 检查是否指定了ID
	if reward.Id == 0 {
		return WithoutIdErr
	}
	// 根据ID更新打赏记录
	if result := dao.db.Where("id = ?", reward.Id).Updates(reward); result.Error != nil {
		return result.Error
	}

	return nil
}

func (dao *RewardDao) UpdateRewardStatus(outTradeNo string, status string) error {
	// 根据ID更新打赏记录
	if result := dao.db.Model(&RewardRecord{}).Where("out_trade_no = ?", outTradeNo).Update("pay_status", status); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		// 如果没有影响的行数，说明不存在或当前状态之前就已被修改为status
		return NoRowsAffectedErr
	}

	return nil
}

func (dao *RewardDao) GetRewardByOutTradeNo(outTradeNo string) (*RewardRecord, error) {
	var reward RewardRecord
	if result := dao.db.Where("out_trade_no = ?", outTradeNo).First(&reward); result.Error != nil {
		return nil, result.Error
	}
	return &reward, nil
}
