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
	UpdateBalance(ctx context.Context, uid int64, accountType domain.AccountType, amount int64, currency string) error
	CreateActivity(ctx context.Context, activity *domain.AccountActivity) error
	GetAccount(ctx context.Context, uid, account int64, accountType domain.AccountType) (*domain.Account, error)
	ProcessCredits(ctx context.Context, biz string, outTradeNo string, creditItems []*domain.CreditItem) error
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
		AccountType: uint8(account.AccountType),
		Balance:     account.Balance,
		Currency:    account.Currency,
	})
}

// UpdateBalance 更新账户余额,账户不存在时会创建账户
func (r *AccountRepository) UpdateBalance(ctx context.Context, uid int64, accountType domain.AccountType, amount int64, currency string) error {
	dao := db.NewAccountDao(ctx, r.GetDB())
	account := &domain.Account{
		Uid:         uid,
		AccountType: accountType,
		Balance:     amount,
		Currency:    currency,
	}

	return dao.UpBalanceOrCreate(AccountDomainToDB(account), account.Balance)

}

func (r *AccountRepository) CreateActivity(ctx context.Context, activity *domain.AccountActivity) error {
	dao := db.NewAccountDao(ctx, r.GetDB())
	return dao.CreateActivity(&db.AccountActivity{
		Uid:         activity.Uid,
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
		AccountType: domain.AccountType(dbAccount.AccountType),
		Balance:     dbAccount.Balance,
		Currency:    dbAccount.Currency,
	}, nil
}

func AccountDomainToDB(account *domain.Account) *db.Account {
	return &db.Account{
		Uid:         account.Uid,
		AccountType: uint8(account.AccountType),
		Balance:     account.Balance,
		Currency:    account.Currency,
	}
}

func (r *AccountRepository) ProcessCredits(ctx context.Context, biz string, outTradeNo string, creditItems []*domain.CreditItem) error {
	return r.GetDB().Transaction(func(tx *gorm.DB) error {
		dao := db.NewAccountDao(ctx, tx)

		for _, item := range creditItems {
			account := &domain.Account{
				Uid:         item.Uid,
				AccountType: item.AccountType,
				Balance:     item.Amount,
				Currency:    item.Currency,
			}
			// 更新账户余额
			if err := dao.UpBalanceOrCreate(AccountDomainToDB(account), item.Amount); err != nil {
				return err
			}

			// 创建账户活动记录
			activity := &db.AccountActivity{
				Uid:         item.Uid,
				AccountType: uint8(item.AccountType),
				Biz:         biz,
				OutTradeNo:  outTradeNo,
				Amount:      item.Amount,
				Currency:    item.Currency,
			}

			if err := dao.CreateActivity(activity); err != nil {
				return err
			}
		}
		return nil
	})
}
