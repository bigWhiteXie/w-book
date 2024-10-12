package dao

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"codexie.com/w-book-common/common/sql"
	"codexie.com/w-book-user/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	dmysql "github.com/go-sql-driver/mysql"
	"github.com/golang/mock/gomock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestUserDao_FindOne(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx  context.Context
		user *model.User
	}
	tests := []struct {
		name    string
		fields  fields
		mock    func(ctrl *gomock.Controller) *UserDao
		args    args
		want    *model.User
		wantErr bool
	}{
		{
			name: "查找用户成功",
			mock: func(ctrl *gomock.Controller) *UserDao {
				db, mock, _ := sqlmock.New()
				rows := sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "125")
				// 转义符不能忘！
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE `users`.`email` = \\? ORDER BY `users`.`id` LIMIT \\?").WithArgs("125", 1).WillReturnRows(rows)
				gormDB, _ := gorm.Open(mysql.New(mysql.Config{
					Conn:                      db, // 将 *sql.DB 传入 gorm
					SkipInitializeWithVersion: true,
				}), &gorm.Config{
					Logger:                 logger.Default.LogMode(logger.Info),
					DisableAutomaticPing:   true,
					SkipDefaultTransaction: true,
				})
				return NewUserDao(gormDB)
			},
			args: args{
				ctx:  context.Background(),
				user: &model.User{Email: sql.StringToNullString("125")},
			},
			want:    &model.User{Id: 1, Email: sql.StringToNullString("125")},
			wantErr: false,
		},
		{
			name: "未查找到用户",
			mock: func(ctrl *gomock.Controller) *UserDao {
				db, mock, _ := sqlmock.New()
				// 转义符不能忘！
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE `users`.`email` = \\? ORDER BY `users`.`id` LIMIT \\?").WillReturnError(gorm.ErrRecordNotFound)
				gormDB, _ := gorm.Open(mysql.New(mysql.Config{
					Conn:                      db, // 将 *sql.DB 传入 gorm
					SkipInitializeWithVersion: true,
				}), &gorm.Config{
					Logger:                 logger.Default.LogMode(logger.Info),
					DisableAutomaticPing:   true,
					SkipDefaultTransaction: true,
				})
				return NewUserDao(gormDB)
			},
			args: args{
				ctx:  context.Background(),
				user: &model.User{Email: sql.StringToNullString("125")},
			},
			want:    nil,
			wantErr: true,
		},
	}
	ctrl := gomock.NewController(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.mock(ctrl)
			got, err := d.FindOne(tt.args.ctx, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserDao.FindOne() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserDao.FindOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserDao_Create(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx  context.Context
		user *model.User
	}
	tests := []struct {
		name    string
		fields  fields
		mock    func(ctrl *gomock.Controller) *UserDao
		args    args
		wantErr bool
	}{
		{
			name: "创建用户成功",
			mock: func(ctrl *gomock.Controller) *UserDao {
				db, mock, _ := sqlmock.New()
				// 转义符不能忘！
				mock.ExpectExec("^INSERT INTO `users` \\(.*\\) VALUES \\(.*\\)").WillReturnResult(sqlmock.NewResult(1, 1))
				gormDB, _ := gorm.Open(mysql.New(mysql.Config{
					Conn:                      db, // 将 *sql.DB 传入 gorm
					SkipInitializeWithVersion: true,
				}), &gorm.Config{
					Logger:                 logger.Default.LogMode(logger.Info),
					DisableAutomaticPing:   true,
					SkipDefaultTransaction: true,
				})
				return NewUserDao(gormDB)
			},
			args: args{
				ctx:  context.Background(),
				user: &model.User{Email: sql.StringToNullString("123"), Password: "123"},
			},
			wantErr: false,
		},
		{
			name: "唯一索引冲突",
			mock: func(ctrl *gomock.Controller) *UserDao {
				db, mock, _ := sqlmock.New()
				// 转义符不能忘！
				mock.ExpectExec("^INSERT INTO `users` \\(.*\\) VALUES \\(.*\\)").WillReturnError(&dmysql.MySQLError{Number: uint16(1062)})
				gormDB, _ := gorm.Open(mysql.New(mysql.Config{
					Conn:                      db, // 将 *sql.DB 传入 gorm
					SkipInitializeWithVersion: true,
				}), &gorm.Config{
					Logger:                 logger.Default.LogMode(logger.Info),
					DisableAutomaticPing:   true,
					SkipDefaultTransaction: true,
				})
				return NewUserDao(gormDB)
			},
			args: args{
				ctx:  context.Background(),
				user: &model.User{Email: sql.StringToNullString("123"), Password: "123"},
			},
			wantErr: true,
		},
		{
			name: "其它异常",
			mock: func(ctrl *gomock.Controller) *UserDao {
				db, mock, _ := sqlmock.New()
				// 转义符不能忘！
				mock.ExpectExec("^INSERT INTO `users` \\(.*\\) VALUES \\(.*\\)").WillReturnError(errors.New("数据库压力过大"))
				gormDB, _ := gorm.Open(mysql.New(mysql.Config{
					Conn:                      db, // 将 *sql.DB 传入 gorm
					SkipInitializeWithVersion: true,
				}), &gorm.Config{
					Logger:                 logger.Default.LogMode(logger.Info),
					DisableAutomaticPing:   true,
					SkipDefaultTransaction: true,
				})
				return NewUserDao(gormDB)
			},
			args: args{
				ctx:  context.Background(),
				user: &model.User{Email: sql.StringToNullString("123"), Password: "123"},
			},
			wantErr: true,
		},
	}
	ctrl := gomock.NewController(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.mock(ctrl)
			if err := d.Create(tt.args.ctx, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("UserDao.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
