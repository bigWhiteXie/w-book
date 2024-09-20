package dao

import (
	"context"
	"reflect"
	"testing"

	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/pkg/common/sql"
	"github.com/DATA-DOG/go-sqlmock"
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
				mock.ExpectQuery("SELECT * FROM `users` WHERE `users`.`email` = ? ORDER BY `users`.`id` LIMIT ?").WithArgs("125", 1).WillReturnRows(rows)
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
			want:    &model.User{Id: 0, Email: sql.StringToNullString("125")},
			wantErr: false,
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
