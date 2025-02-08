package db

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Account 账户表
type Account struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	Uid         int64 `gorm:"uniqueIndex:idx_uid_accounttype"`
	AccountType uint8 `gorm:"uniqueIndex:idx_uid_accounttype"`
	Balance     int64
	Currency    string
	Ctime       int64
	Utime       int64
}

// AccountActivity 账户活动表
type AccountActivity struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	Uid         int64  `gorm:"index:idx_uid_account_type,priority:1"`
	AccountType uint8  `gorm:"index:idx_uid_account_type,priority:2"`
	Biz         string `gorm:"index:idx_biz_trade_no,priority:1,length:64"`
	OutTradeNo  string `gorm:"index:idx_biz_trade_no,priority:2,length:64"`
	Amount      int64
	Currency    string
	Ctime       int64
	Utime       int64
}

type AccountDao struct {
	db  *gorm.DB
	log logx.Logger
	ctx context.Context
}

func NewAccountDao(ctx context.Context, db *gorm.DB) *AccountDao {
	return &AccountDao{
		db:  db,
		log: logx.WithContext(ctx),
		ctx: ctx,
	}
}

// CreateAccount 创建账户
func (dao *AccountDao) CreateAccount(account *Account) error {
	now := time.Now().Unix()
	account.Ctime = now
	account.Utime = now

	result := dao.db.Create(account)
	return result.Error
}

// UpdateBalance 更新账户余额
func (dao *AccountDao) UpdateBalance(uid int64, accountType uint8, amount int64) error {
	sql := dao.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&Account{}).
			Where("uid = ? AND account_type = ?", uid, accountType).
			UpdateColumn("balance", gorm.Expr("`balance` + ?", amount))
	})

	dao.log.Infof("Generated SQL: %s", sql)
	result := dao.db.Model(&Account{}).
		Where("uid = ? AND account_type = ?", uid, accountType).
		UpdateColumn("balance", gorm.Expr("balance + ?", amount))
	return result.Error
}

// CreateActivity 创建账户活动记录
func (dao *AccountDao) CreateActivity(activity *AccountActivity) error {
	now := time.Now().Unix()
	activity.Ctime = now
	activity.Utime = now

	result := dao.db.Create(activity)
	return errors.Wrap(result.Error, "创建账户活动记录失败")
}

// GetAccount 获取账户信息
func (dao *AccountDao) GetAccount(uid, account int64, accountType uint8) (*Account, error) {
	var acc Account
	result := dao.db.Where("uid = ? AND account = ? AND account_type = ?", uid, account, accountType).
		First(&acc)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "获取账户信息失败")
	}
	return &acc, nil
}

// UpBalanceOrCreate 更新余额或创建账户
func (dao *AccountDao) UpBalanceOrCreate(account *Account, amount int64) error {
	now := time.Now().UnixMilli()
	account.Ctime = now
	account.Utime = now

	if err := dao.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "uid"}, {Name: "account_type"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"balance": gorm.Expr("`balance` + ?", amount),
			"utime":   now,
		}),
	}).Create(account).Error; err != nil {
		return errors.Wrap(err, "更新余额或创建账户失败")
	}
	return nil
}
