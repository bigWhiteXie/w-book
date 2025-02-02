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

func FromReward(reward *db.RewardRecord) *domain.RewardRecord {
	return &domain.RewardRecord{
		Id:         reward.Id,
		Uid:        reward.Uid,
		Biz:        reward.Biz,
		ResourceId: reward.ResourceId,
		Status:     reward.Status,
		OutTradeNo: reward.OutTradeNo,
	}
}

func ToRewardEntity(reward *domain.RewardRecord) *db.RewardRecord {
	now := time.Now().Unix()
	return &db.RewardRecord{
		Id:         reward.Id,
		Biz:        reward.Biz,
		ResourceId: reward.ResourceId,
		Status:     reward.Status,
		OutTradeNo: reward.OutTradeNo,
		Ctime:      now,
		Utime:      now,
	}
}
