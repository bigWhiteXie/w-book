package dao

import (
	"codexie.com/w-book-user/internal/common"
	"codexie.com/w-book-user/internal/model"
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{db: db}
}

// dao无需考虑isTx是因为dao的方法中不会调用Tx方法
// Tx方法是为上层提供事务的
func (d *UserDao) TX(fun func(dao *UserDao) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		txDao := NewUserDao(tx)
		return fun(txDao)
	})
}

func (d *UserDao) Create(ctx context.Context, user *model.User) error {
	err := d.db.WithContext(ctx).Create(user).Error
	user.Utime = time.Now()
	user.Utime = time.Now()
	if me, ok := err.(*mysql.MySQLError); ok {
		if uniqueIndexErrNo := uint16(1062); me.Number == uniqueIndexErrNo {
			return common.UserEmailDuplicateErr
		}
	}
	return err
}

func (d *UserDao) Find(ctx context.Context, user *model.User) (users []model.User, err error) {
	err = d.db.WithContext(ctx).Where(user).Find(&users).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.UserEmailNotExistErr
		}
		return nil, err
	}
	return
}
