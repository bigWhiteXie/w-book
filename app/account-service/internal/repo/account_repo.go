package repo

import (
	"context"

	"codexie.com/w-book-account/internal/dao/db"
	"codexie.com/w-book-account/internal/domain"
	"codexie.com/w-book-common/repo"
	"gorm.io/gorm"
)

type IAccountRepository interface {
	CreateAccount(ctx context.Context, account *domain.Account) error
	UpdateBalance(ctx context.Context, uid, account int64, accountType domain.AccountType, amount int64) error
	CreateActivity(ctx context.Context, activity *domain.AccountActivity) error
	GetAccount(ctx context.Context, uid, account int64, accountType domain.AccountType) (*domain.Account, error)
}

type AccountRepository struct {
	*repo.BaseRepo
}

func NewAccountRepository(gormDB *gorm.DB) *AccountRepository {
	if err := gormDB.AutoMigrate(&db.Account{}, &db.AccountActivity{}); err != nil {
		panic(err)
	}
	return &AccountRepository{
		repo.NewBaseRepo(gormDB),
	}
}

func (r *AccountRepository) CreateAccount(ctx context.Context, account *domain.Account) error {
	dao := db.NewAccountDao(ctx, r.GetDB())
	return dao.CreateAccount(&db.Account{
		Uid:         account.Uid,
		Account:     account.Account,
		AccountType: uint8(account.AccountType),
		Balance:     account.Balance,
		Currency:    account.Currency,
	})
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, uid, account int64, accountType domain.AccountType, amount int64) error {
	dao := db.NewAccountDao(ctx, r.GetDB())

	return dao.UpBalanceOrCreate(&db.Account{
		Uid:         uid,
		Account:     account,
		AccountType: uint8(accountType),
	}, amount)
}

func (r *AccountRepository) CreateActivity(ctx context.Context, activity *domain.AccountActivity) error {
	dao := db.NewAccountDao(ctx, r.GetDB())
	return dao.CreateActivity(&db.AccountActivity{
		Uid:         activity.Uid,
		Account:     activity.Account,
		AccountType: uint8(activity.AccountType),
		Biz:         activity.Biz,
		OutTradeNo:  activity.OutTradeNo,
		Amount:      activity.Amount,
		Currency:    activity.Currency,
	})
}

func (r *AccountRepository) GetAccount(ctx context.Context, uid, account int64, accountType domain.AccountType) (*domain.Account, error) {
	dao := db.NewAccountDao(ctx, r.GetDB())
	dbAccount, err := dao.GetAccount(uid, account, uint8(accountType))
	if err != nil {
		return nil, err
	}

	return &domain.Account{
		Id:          dbAccount.Id,
		Uid:         dbAccount.Uid,
		Account:     dbAccount.Account,
		AccountType: domain.AccountType(dbAccount.AccountType),
		Balance:     dbAccount.Balance,
		Currency:    dbAccount.Currency,
	}, nil
}
