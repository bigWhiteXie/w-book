package db

import (
	"context"

	"time"

	"codexie.com/w-book-payment/pkg/constant"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

var (
	NoRowsAffectedErr = errors.New("no rows affected")
)

type PaymentRecord struct {
	Id         int64  `json:"", gorm:"primaryKey,autoIncrement"`
	Biz        string `json:"", gorm:"uniqueIndex:idx_biz_tradeNo"`
	OutTradeNo string `json:"", gorm:"uniqueIndex:idx_biz_tradeNo"`
	Platform   string `json:""`
	Status     string `json:"", gorm:"index:idx_status_utime"`
	Amt        int64  `json:""`
	Currency   string `json:""`
	Subject    string `json:""`
	Ctime      int64  `json:""`
	Utime      int64  `json:"", gorm:"index:idx_status_utime"`
}

type PaymentDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	log logx.Logger
	db  *gorm.DB
	ctx context.Context
}

func NewPaymentDao(ctx context.Context, gormDB *gorm.DB) *PaymentDao {

	return &PaymentDao{
		db:  gormDB,
		log: logx.WithContext(ctx),
		ctx: ctx,
	}
}

func (dao *PaymentDao) UpdateStatusByBizAndOutTradeNo(biz, outTradeNo, status string) error {
	result := dao.db.Model(&PaymentRecord{}).
		Where("biz = ? AND out_trade_no = ?", biz, outTradeNo).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return NoRowsAffectedErr
	}

	return nil
}

func (dao *PaymentDao) FindInitPaymentsByLastId(lastId int64, limit int) ([]*PaymentRecord, error) {
	var payments []*PaymentRecord
	currentTime := time.Now().Unix()
	result := dao.db.Where("status = ? AND ctime < ? AND id > ?", constant.InitPayStatus, currentTime-30, lastId).
		Limit(limit).
		Find(&payments)
	if result.Error != nil {
		return nil, result.Error
	}
	return payments, nil
}

func (dao *PaymentDao) CreatePaymentRecord(record *PaymentRecord) error {
	result := dao.db.Create(&record)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (dao *PaymentDao) UpdateStatusByIds(ids []int64, status string) error {
	result := dao.db.WithContext(dao.ctx).Where("id in (?)", ids).Update("status", status)
	if result.Error != nil {
		return errors.Wrap(result.Error, "更新payment状态异常")
	}

	return nil
}

func (dao *PaymentDao) FindPaymentByBiz(biz string, outTradeNo string) (*PaymentRecord, error) {
	record := &PaymentRecord{}
	result := dao.db.
		Where("biz = ? AND out_trade_no = ?", biz, outTradeNo).
		First(record)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "查询payment异常")
	}

	return record, nil
}
