package repo

import (
	"context"
	"time"

	"codexie.com/w-book-common/repo"
	"codexie.com/w-book-reward/internal/dao/db"
	"codexie.com/w-book-reward/internal/domain"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type IRewardRepository interface {
	CreateReward(ctx context.Context, reward *domain.RewardRecord) (int64, error)
	UpdateReward(ctx context.Context, reward *domain.RewardRecord) error
	GetRewardByOutTradeNo(ctx context.Context, outTradeNo string) (*domain.RewardRecord, error)
	UpStatusByOutTradeNo(ctx context.Context, outTradeNo string, status string) error
}

type RewardRepository struct {
	*repo.BaseRepo
	g    singleflight.Group
	isTx bool
}

func NewRewardRepository(gormDB *gorm.DB) *RewardRepository {
	if err := gormDB.AutoMigrate(&db.RewardRecord{}); err != nil {
		panic(err)
	}

	return &RewardRepository{BaseRepo: repo.NewBaseRepo(gormDB)}
}

func (r *RewardRepository) CreateReward(ctx context.Context, reward *domain.RewardRecord) (int64, error) {
	rewardDao := db.NewRewardDao(ctx, r.GetDB())
	return rewardDao.CreateReward(ToRewardEntity(reward))
}

func (r *RewardRepository) UpdateReward(ctx context.Context, reward *domain.RewardRecord) error {
	rewardDao := db.NewRewardDao(ctx, r.GetDB())
	return rewardDao.UpdateRewardById(ToRewardEntity(reward))
}

func (r *RewardRepository) UpStatusByOutTradeNo(ctx context.Context, outTradeNo string, status string) error {
	rewardDao := db.NewRewardDao(ctx, r.GetDB())
	return rewardDao.UpdateRewardStatus(outTradeNo, status)
}

// 根据outTradeNo获得打赏记录
func (r *RewardRepository) GetRewardByOutTradeNo(ctx context.Context, outTradeNo string) (*domain.RewardRecord, error) {
	rewardDao := db.NewRewardDao(ctx, r.GetDB())
	reward, err := rewardDao.GetRewardByOutTradeNo(outTradeNo)
	if err != nil {
		return nil, err
	}
	return FromReward(reward), nil
}

func FromReward(reward *db.RewardRecord) *domain.RewardRecord {
	return &domain.RewardRecord{
		Id:         reward.Id,
		Biz:        reward.Biz,
		Uid:        reward.Uid,
		ResourceId: reward.ResourceId,
		AuthorId:   reward.AuthorId,
		Status:     reward.PayStatus,
		OutTradeNo: reward.OutTradeNo,
		Amt:        reward.Amt,
		Platform:   reward.Platform,
	}
}

func ToRewardEntity(reward *domain.RewardRecord) *db.RewardRecord {
	now := time.Now().Unix()
	return &db.RewardRecord{
		Id:         reward.Id,
		Biz:        reward.Biz,
		Uid:        reward.Uid,
		ResourceId: reward.ResourceId,
		AuthorId:   reward.AuthorId,
		PayStatus:  reward.Status,
		OutTradeNo: reward.OutTradeNo,
		Amt:        reward.Amt,
		Platform:   reward.Platform,
		Ctime:      now,
		Utime:      now,
	}
}
