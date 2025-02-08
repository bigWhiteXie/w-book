package service

import (
	"context"

	"codexie.com/w-book-account/api/pb"
	"codexie.com/w-book-account/internal/domain"
	"codexie.com/w-book-account/internal/repo"
)

type AccountService struct {
	pb.UnimplementedAccountServer
	accountRepo repo.IAccountRepository
}

func NewAccountService(accountRepo repo.IAccountRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
	}
}

func (s *AccountService) Credit(ctx context.Context, req *pb.CreditRequest) (*pb.CreditResponse, error) {
	creditItems := make([]*domain.CreditItem, len(req.CreditItems))
	for i, item := range req.CreditItems {
		creditItems[i] = &domain.CreditItem{
			Uid:         item.Uid,
			AccountType: domain.AccountType(item.AccountType),
			Amount:      item.Amt,
			Currency:    item.Currency,
		}
	}

	if err := s.accountRepo.ProcessCredits(ctx, req.Biz, req.OutTradeNo, creditItems); err != nil {
		return nil, err
	}

	return &pb.CreditResponse{
		Code:    0,
		Message: "Success",
	}, nil
}
