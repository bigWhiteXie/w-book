package dao

import (
	"context"
	"strconv"

	"codexie.com/w-book-code/internal/model"
	"codexie.com/w-book-common/common/codeerr"

	"gorm.io/gorm"
)

type CodeDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewCodeDao(db *gorm.DB) *CodeDao {
	return &CodeDao{db: db}
}

// dao无需考虑isTx是因为dao的方法中不会调用Tx方法
// Tx方法是为上层提供事务的
func (d *CodeDao) TX(fun func(dao *CodeDao) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		txDao := NewCodeDao(tx)
		return fun(txDao)
	})
}

func (d *CodeDao) Save(ctx context.Context, record *model.SmsSendRecord) error {
	err := d.db.WithContext(ctx).Create(record).Error
	return err
}

func (d *CodeDao) Update(ctx context.Context, record *model.SmsSendRecord) error {
	err := d.db.WithContext(ctx).Save(record).Error
	return err
}

func (d *CodeDao) FindById(ctx context.Context, idstr string) (*model.SmsSendRecord, error) {
	id, _ := strconv.Atoi(idstr)
	record := &model.SmsSendRecord{}
	if err := d.db.WithContext(ctx).First(record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, codeerr.WithCode(codeerr.SmsNotFoundErr, "fail to find record by id:%s", id)
		}
		return nil, err
	}
	return record, nil
}
