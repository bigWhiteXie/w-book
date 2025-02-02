package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 定义测试用例
func TestPayMsgRepo_CreateMsg(t *testing.T) {
	// 初始化 sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// 使用 sqlmock 初始化 GORM DB
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm DB: %v", err)
	}
	mock.ExpectExec("CREATE TABLE `pay_msgs`").
		WithArgs().                               // 如果有特定的参数传递，可以在这里填写
		WillReturnResult(sqlmock.NewResult(0, 0)) // 创建表的操作通常不返回 ID，所以传递 0, 0
	// 初始化 PayMsgRepo
	repo := NewPayMsgRepo(gormDB)

	// 定义测试用例
	tests := []struct {
		name        string
		topic       string
		msg         string
		mockExpect  func() // 模拟数据库行为
		wantErr     bool   // 是否期望错误
		expectedErr error  // 期望的错误
	}{
		{
			name:  "Success - Message created",
			topic: "test-topic",
			msg:   "test-message",
			mockExpect: func() {
				// 模拟数据库插入成功
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `pay_msgs`").
					WithArgs("test-topic", "test-message", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:  "Failure - Database error",
			topic: "test-topic",
			msg:   "test-message",
			mockExpect: func() {
				// 模拟数据库插入失败
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `pay_msgs`").
					WithArgs("test-topic", "test-message", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
	}

	// 遍历测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟行为
			tt.mockExpect()

			// 调用被测方法
			err := repo.CreateMsg(context.Background(), tt.topic, tt.msg)

			// 断言结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			// 确保所有期望的数据库操作都被执行
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
