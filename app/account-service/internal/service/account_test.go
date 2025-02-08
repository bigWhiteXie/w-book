package service

import (
	"context"
	"testing"

	"codexie.com/w-book-account/api/pb"
	"codexie.com/w-book-account/internal/config"
	"codexie.com/w-book-account/internal/dao/db"
	"codexie.com/w-book-account/internal/domain"
	"codexie.com/w-book-account/internal/repo"
	"codexie.com/w-book-common/ioc"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/conf"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	var c config.Config
	conf.MustLoad("/usr/local/go_project/w-book/app/account-service/etc/account.yaml", &c)
	db := ioc.InitGormDB(c.MySQLConf)
	return db
}

func TestAccountService_Credit(t *testing.T) {
	gormDB := setupTestDB()
	accountRepo := repo.NewAccountRepository(gormDB)
	service := NewAccountService(accountRepo)

	tests := []struct {
		name    string
		req     *pb.CreditRequest
		wantErr bool
		setup   func()
		verify  func(t *testing.T)
	}{
		{
			name: "successful credit",
			req: &pb.CreditRequest{
				Biz:        "test",
				OutTradeNo: "test-123",
				CreditItems: []*pb.CreditItem{
					{
						Uid:         2,
						AccountType: pb.AccountType_Reward,
						Amt:         900,
						Currency:    "CNY",
					},
					{
						Uid:         0,
						AccountType: pb.AccountType_Reward,
						Amt:         100,
						Currency:    "CNY",
					},
				},
			},
			setup: func() {
				// 清理测试数据
				gormDB.Exec("DELETE FROM account")
				gormDB.Exec("DELETE FROM account_activity")
			},
			verify: func(t *testing.T) {
				// 验证账户余额
				var account db.Account
				gormDB.Where("uid = ? AND account_type = ?", 2, domain.AccountTypeReward).First(&account)
				assert.Equal(t, int64(900), account.Balance)

				// 验证活动记录
				var activity db.AccountActivity
				gormDB.Where("uid = ? AND account_type = ?", 0, domain.AccountTypeReward).First(&activity)
				assert.Equal(t, int64(100), activity.Amount)
			},
		},
		//再构建一个测试用例，outtradeNo设置为超出数据库限制的长度验证事务回滚
		{
			name: "invalid currency",
			// setup: func() {
			// 	// 清理测试数据
			// 	gormDB.Exec("DELETE FROM account")
			// 	gormDB.Exec("DELETE FROM account_activity")
			// },
			req: &pb.CreditRequest{
				OutTradeNo: "12345",
				Biz:        "test-1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
				CreditItems: []*pb.CreditItem{
					{
						Uid:         2,
						AccountType: pb.AccountType_Reward,
						Amt:         900,
						Currency:    "CNY",
					},
					{
						Uid:         0,
						AccountType: pb.AccountType_Reward,
						Amt:         100,
						Currency:    "CNY",
					},
				},
			},
			wantErr: true,
		},
		// 添加更多测试用例...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			_, err := service.Credit(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}
