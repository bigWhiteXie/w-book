package db

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // 打印日志到标准输出
			logger.Config{
				SlowThreshold:        time.Second, // 慢查询的阈值
				LogLevel:             logger.Info, // 日志级别
				ParameterizedQueries: true,
				Colorful:             true, // 启用彩色输出
			},
		),
	})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}

	return gormDB, mock
}

func TestPayMsgDao_FindByLastId(t *testing.T) {
	db, mock := setupMockDB(t)
	dao := NewPayMsgDao(context.Background(), db)

	tests := []struct {
		name     string
		lastId   int64
		limit    int
		mockFunc func()
		want     []PayMsg
		wantErr  bool
	}{
		{
			name:   "successful query",
			lastId: 1,
			limit:  10,
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "topic", "msg", "retry_times", "ctime", "utime"}).
					AddRow(2, "topic1", "msg1", 0, 1, 1).
					AddRow(3, "topic2", "msg2", 0, 2, 2)
				mock.ExpectQuery("^SELECT \\* FROM `pay_msgs` WHERE id > \\? ORDER BY id ASC LIMIT \\?$").
					WithArgs(1, 10).
					WillReturnRows(rows)
			},
			want: []PayMsg{
				{Id: 2, Topic: "topic1", Msg: "msg1", RetryTimes: 0, Ctime: 1, Utime: 1},
				{Id: 3, Topic: "topic2", Msg: "msg2", RetryTimes: 0, Ctime: 2, Utime: 2},
			},
			wantErr: false,
		},
		{
			name:   "query error",
			lastId: 1,
			limit:  10,
			mockFunc: func() {
				mock.ExpectQuery("^SELECT \\* FROM `pay_msgs` WHERE id > \\? ORDER BY id ASC LIMIT \\?$").
					WithArgs(1, 10).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			got, err := dao.FindByLastId(tt.lastId, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("PayMsgDao.FindByLastId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPayMsgDao_DeleteByIds(t *testing.T) {
	db, mock := setupMockDB(t)
	dao := NewPayMsgDao(context.Background(), db)

	tests := []struct {
		name     string
		ids      []int64
		mockFunc func()
		wantErr  bool
	}{
		{
			name: "successful delete",
			ids:  []int64{1, 2, 3},
			mockFunc: func() {
				mock.ExpectBegin() // 显式模拟 Begin 操作
				mock.ExpectExec("^DELETE FROM `pay_msgs` WHERE id IN \\(\\?,\\?,\\?\\)$").
					WithArgs(1, 2, 3).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "delete error",
			ids:  []int64{1, 2, 3},
			mockFunc: func() {
				mock.ExpectBegin() // 显式模拟 Begin 操作
				mock.ExpectExec("^DELETE FROM `pay_msgs` WHERE id IN \\(\\?,\\?,\\?\\)$").
					WithArgs(1, 2, 3).
					WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectCommit()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			err := dao.DeleteByIds(tt.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("PayMsgDao.DeleteByIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPayMsgDao_CreatePayMsg(t *testing.T) {
	db, mock := setupMockDB(t)
	dao := NewPayMsgDao(context.Background(), db)

	tests := []struct {
		name     string
		topic    string
		msg      string
		mockFunc func()
		wantErr  bool
	}{
		{
			name:  "successful create",
			topic: "topic1",
			msg:   "msg1",
			mockFunc: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO `pay_msgs` \\(`topic`,`msg`,`retry_times`,`ctime`,`utime`\\) VALUES \\(\\?,\\?,\\?,\\?,\\?\\)$").
					WithArgs("topic1", "msg1", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:  "create error",
			topic: "topic1",
			msg:   "msg1",
			mockFunc: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO `pay_msgs` \\(`topic`,`msg`,`retry_times`,`ctime`,`utime`\\) VALUES \\(\\?,\\?,\\?,\\?,\\?\\)$").
					WithArgs("topic1", "msg1", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			err := dao.CreatePayMsg(tt.topic, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("PayMsgDao.CreatePayMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
