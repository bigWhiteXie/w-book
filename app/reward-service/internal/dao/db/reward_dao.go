package db

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

var (
	WithoutIdErr  = errors.New("id must specified")
	IdNotExistErr = errors.New("reward id is not exist")
)

type RewardRecord struct {
	Id         int64  `json:"",gorm:"primaryKey,autoIncrement"`        // 主键，自增
	Uid        int64  `json:""`                                        // 打赏人ID
	AuthorId   int64  `json:""`                                        // 作者ID
	Amt        int64  `json:""`                                        // 金额
	Biz        string `json:"", gorm:"uniqueIndex:idx_biz_resourceId"` // 资源类型，唯一索引
	ResourceId int64  `json:"", gorm:"uniqueIndex:idx_biz_resourcrId"` // 资源ID，唯一索引
	OutTradeNo string `json:"", gorm:"uniqueIndex:idx_tradeNo"`        // 外贸编号，唯一索引
	Platform   string `json:""`                                        // 平台
	Status     string `json:""`                                        // 状态
	Ctime      int64  `json:""`                                        // 创建时间
	Utime      int64  `json:"",gorm:"index:idx_uid_uptime"`            // 更新时间，索引
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

func (dao *RewardDao) UpdateRewardById(reward *RewardRecord) error {
	if reward.Id == 0 {
		return WithoutIdErr
	}
	if result := dao.db.Where("id = ?", reward.Id).Updates(reward); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return IdNotExistErr
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
