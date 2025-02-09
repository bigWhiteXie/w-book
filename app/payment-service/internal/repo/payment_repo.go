package repo

import (
	"context"
	"time"

	"codexie.com/w-book-common/repo"
	"codexie.com/w-book-payment/internal/dao/db"
	"codexie.com/w-book-payment/internal/domain"
	"codexie.com/w-book-payment/pkg/constant"
	"gorm.io/gorm"
)

type IPaymentRepository interface {
	CreatePayment(ctx context.Context, payment *domain.Payment) error
	UpdatePaymentStatus(ctx context.Context, biz string, outTradeNo string, status constant.PayStatus) error
	FindPaymentByBizAndOutTradeNo(ctx context.Context, biz string, outTradeNo string) (*domain.Payment, error)
	FindInitPaymentsByCtime(ctx context.Context, minCtime, maxCtime int64, limit int) ([]*domain.Payment, error)
	UpdateStatusByIds(ctx context.Context, ids []int64, status constant.PayStatus) error
}

type PaymentRepository struct {
	*repo.BaseRepo
}

func NewPaymentRepository(gormDB *gorm.DB) *PaymentRepository {
	if err := gormDB.AutoMigrate(&db.PaymentRecord{}); err != nil {
		panic(err)
	}

	return &PaymentRepository{
		repo.NewBaseRepo(gormDB),
	}
}

// CreatePayment 创建支付记录
func (r *PaymentRepository) CreatePayment(ctx context.Context, payment *domain.Payment) error {
	// 将 domain.Payment 转换为 db.PaymentRecord
	record := &db.PaymentRecord{
		Platform:   payment.Platform,
		Biz:        payment.Biz,
		OutTradeNo: payment.OutTradeNo,
		Subject:    payment.Subject,
		Amt:        payment.Amt,
		Currency:   payment.Currency,
		Status:     string(payment.Status),
		Ctime:      time.Now().Unix(), // 设置创建时间
		Utime:      time.Now().Unix(), // 设置更新时间
	}

	payDao := db.NewPaymentDao(ctx, r.GetDB())
	return payDao.CreatePaymentRecord(record)
}

// UpdatePaymentStatus 更新支付状态
func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, biz string, outTradeNo string, status constant.PayStatus) error {
	payDao := db.NewPaymentDao(ctx, r.GetDB())
	err := payDao.UpdateStatusByBizAndOutTradeNo(biz, outTradeNo, string(status))
	if err != nil {
		return err
	}
	return nil
}

// FindPaymentByBizAndOutTradeNo 根据 biz 和 outTradeNo 查找支付记录
func (r *PaymentRepository) FindPaymentByBizAndOutTradeNo(ctx context.Context, biz string, outTradeNo string) (*domain.Payment, error) {
	payDao := db.NewPaymentDao(ctx, r.GetDB())
	record, err := payDao.FindPaymentByBiz(biz, outTradeNo)
	if err != nil {
		return nil, err
	}

	return paymentRecordToDomain(record), nil
}

func (r *PaymentRepository) FindInitPaymentsByCtime(ctx context.Context, minCtime, maxCtime int64, limit int) ([]*domain.Payment, error) {
	var payments []*domain.Payment
	payDao := db.NewPaymentDao(ctx, r.GetDB())
	records, err := payDao.FindInitPaymentsByLastId(limit, minCtime, maxCtime)
	if err != nil {
		return nil, err
	}
	payments = make([]*domain.Payment, 0, len(records))
	for _, record := range records {
		payments = append(payments, paymentRecordToDomain(record))
	}

	return payments, nil
}

func (r *PaymentRepository) UpdateStatusByIds(ctx context.Context, ids []int64, status constant.PayStatus) error {
	dao := db.NewPaymentDao(ctx, r.GetDB())
	return dao.UpdateStatusByIds(ids, string(status))
}

func paymentRecordToDomain(record *db.PaymentRecord) *domain.Payment {
	return &domain.Payment{
		Id:         record.Id,
		Platform:   record.Platform,
		Biz:        record.Biz,
		OutTradeNo: record.OutTradeNo,
		Subject:    record.Subject,
		Amt:        record.Amt,
		Currency:   record.Currency,
		Status:     constant.PayStatus(record.Status),
		Ctime:      record.Ctime,
		Utime:      record.Utime,
	}
}

func paymentDomainToRecord(payment *domain.Payment) *db.PaymentRecord {
	return &db.PaymentRecord{
		Platform:   payment.Platform,
		Biz:        payment.Biz,
		OutTradeNo: payment.OutTradeNo,
		Subject:    payment.Subject,
		Amt:        payment.Amt,
		Currency:   payment.Currency,
		Status:     string(payment.Status),
		Ctime:      payment.Ctime,
		Utime:      payment.Utime,
	}
}
