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
	Id         int64  `json:"",gorm:"primaryKey,autoIncrement"`
	Uid        int64  `json:""`
	Biz        string `json:"", gorm:"uniqueIndex:idx_biz_resourceId"`
	ResourceId int64  `json:"", gorm:"uniqueIndex:idx_biz_resourcrId"`
	OutTradeNo string `json:"", gorm:"uniqueIndex:idx_tradeNo"`
	Platform   string `json:""`
	Status     string `json:""`
	Ctime      int64  `json:""`
	Utime      int64  `json:"",gorm:"index:idx_uid_uptime"`
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
