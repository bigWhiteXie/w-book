package db

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, error) {
	var err error
	var sqlDB *sql.DB
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})

	return gormDB, mock, err
}

func TestAccountDao_CreateAccount(t *testing.T) {
	gormDB, mock, err := setupTestDB(t)
	assert.NoError(t, err)

	dao := NewAccountDao(context.Background(), gormDB)

	tests := []struct {
		name    string
		account *Account
		mock    func()
		wantErr bool
	}{
		{
			name: "success",
			account: &Account{
				Uid:         1,
				AccountType: 1,
				Balance:     100,
				Currency:    "CNY",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `accounts`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "duplicate error",
			account: &Account{
				Uid:         1,
				AccountType: 1,
				Balance:     100,
				Currency:    "CNY",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `accounts`").
					WillReturnError(gorm.ErrDuplicatedKey)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := dao.CreateAccount(tt.account)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAccountDao_UpdateBalance(t *testing.T) {
	gormDB, mock, err := setupTestDB(t)
	assert.NoError(t, err)

	dao := NewAccountDao(context.Background(), gormDB)

	tests := []struct {
		name        string
		uid         int64
		accountType uint8
		amount      int64
		mock        func()
		wantErr     bool
	}{
		{
			name:        "success - positive amount",
			uid:         1,
			accountType: 1,
			amount:      100,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `accounts` SET `balance`=balance \\+ \\? WHERE uid = \\? AND account_type = \\?").
					WithArgs(100, 1, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:        "success - negative amount",
			uid:         1,
			accountType: 1,
			amount:      -50,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `accounts` SET `balance`=balance \\+ \\? WHERE uid = \\? AND account_type = \\?").
					WithArgs(-50, 1, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:        "account not found",
			uid:         999,
			accountType: 1,
			amount:      100,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `accounts` SET `balance`=balance \\+ \\? WHERE uid = \\? AND account_type = \\?").
					WithArgs(100, 999, 1).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:        "database error",
			uid:         1,
			accountType: 1,
			amount:      100,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `accounts` SET `balance`=balance \\+ \\? WHERE uid = \\? AND account_type = \\?").
					WithArgs(100, 1, 1).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := dao.UpdateBalance(tt.uid, tt.accountType, tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 检查是否所有的预期都被满足
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("存在未满足的预期: %s", err)
			}
		})
	}
}

func TestAccountDao_UpBalanceOrCreate(t *testing.T) {
	gormDB, mock, err := setupTestDB(t)
	assert.NoError(t, err)

	dao := NewAccountDao(context.Background(), gormDB)

	tests := []struct {
		name    string
		account *Account
		amount  int64
		mock    func()
		wantErr bool
	}{
		{
			name: "create new account",
			account: &Account{
				Uid:         1,
				AccountType: 1,
				Balance:     100,
				Currency:    "CNY",
			},
			amount: 100,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `accounts` (`uid`,`account_type`,`balance`,`currency`,`ctime`,`utime`) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `balance`=`balance` + ?,`utime`=?")).
					WithArgs(1, 1, 100, "CNY", sqlmock.AnyArg(), sqlmock.AnyArg(), 100, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "update existing account",
			account: &Account{
				Uid:         1,
				AccountType: 1,
				Balance:     100,
				Currency:    "CNY",
			},
			amount: 50,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `accounts` (`uid`,`account_type`,`balance`,`currency`,`ctime`,`utime`) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `balance`=`balance` + ?,`utime`=?")).
					WithArgs(1, 1, 100, "CNY", sqlmock.AnyArg(), sqlmock.AnyArg(), 50, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "database error on create",
			account: &Account{
				Uid:         1,
				AccountType: 1,
				Balance:     100,
				Currency:    "CNY",
			},
			amount: 100,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `accounts` (`uid`,`account_type`,`balance`,`currency`,`ctime`,`utime`) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `balance`=`balance` + ?,`utime`=?")).
					WithArgs(1, 1, 100, "CNY", sqlmock.AnyArg(), sqlmock.AnyArg(), 100, sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "database error on update",
			account: &Account{
				Uid:         1,
				AccountType: 1,
				Balance:     100,
				Currency:    "CNY",
			},
			amount: 50,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `accounts` (`uid`,`account_type`,`balance`,`currency`,`ctime`,`utime`) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `balance`=`balance` + ?,`utime`=?")).
					WithArgs(1, 1, 100, "CNY", sqlmock.AnyArg(), sqlmock.AnyArg(), 50, sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := dao.UpBalanceOrCreate(tt.account, tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 检查是否所有的预期都被满足
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("存在未满足的预期: %s", err)
			}
		})
	}
}

func TestAccountDao_CreateActivity(t *testing.T) {
	gormDB, mock, err := setupTestDB(t)
	assert.NoError(t, err)

	dao := NewAccountDao(context.Background(), gormDB)

	tests := []struct {
		name     string
		activity *AccountActivity
		mock     func()
		wantErr  bool
	}{
		{
			name: "success",
			activity: &AccountActivity{
				Uid:         1,
				AccountType: 1,
				Biz:         "test",
				OutTradeNo:  "123456",
				Amount:      100,
				Currency:    "CNY",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `account_activities`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "duplicate trade no",
			activity: &AccountActivity{
				Uid:         1,
				AccountType: 1,
				Biz:         "test",
				OutTradeNo:  "123456",
				Amount:      100,
				Currency:    "CNY",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `account_activities`").
					WillReturnError(gorm.ErrDuplicatedKey)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := dao.CreateActivity(tt.activity)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
