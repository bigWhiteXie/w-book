package service

import (
	"context"

	"codexie.com/w-book-account/api/pb"
	"codexie.com/w-book-account/internal/domain"
	"codexie.com/w-book-account/internal/repo"
	"gorm.io/gorm"
)

type AccountService struct {
	pb.UnimplementedAccountServer
	db *gorm.DB
}

func NewAccountService(db *gorm.DB) *AccountService {
	return &AccountService{
		db: db,
	}
}

func (s *AccountService) Credit(ctx context.Context, req *pb.CreditRequest) (*pb.CreditResponse, error) {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		accountRepo := repo.NewAccountRepository(tx)

		for _, item := range req.CreditItems {
			accountType := domain.GetAccountType(item.AccountType)
			// 更新账户余额
			err := accountRepo.UpdateBalance(ctx, item.Uid, item.Account, accountType, item.Amt)
			if err != nil {
				return err
			}

			// 创建账户活动记录
			activity := &domain.AccountActivity{
				Uid:         item.Uid,
				Account:     item.Account,
				AccountType: accountType,
				Biz:         req.Biz,
				OutTradeNo:  req.OutTradeNo,
				Amount:      item.Amt,
				Currency:    item.Currency,
			}

			err = accountRepo.CreateActivity(ctx, activity)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &pb.CreditResponse{
		Code:    0,
		Message: "Success",
	}, nil
}
