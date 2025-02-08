package logic

import (
	"context"
	"net/http"
	"strconv"

	"codexie.com/w-book-payment/api/pb"
	"codexie.com/w-book-payment/internal/domain"
	"codexie.com/w-book-payment/internal/repo"
	"codexie.com/w-book-payment/internal/types"
	"codexie.com/w-book-payment/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
)

type PayLogic interface {
	PayCallback(ctx context.Context, req *types.AliPaymentMsg) error
	GetPayment(ctx context.Context, in *pb.QueryPaymentReq) (*pb.QueryPaymentResp, error)
	PrePay(ctx context.Context, in *pb.PrepayReq) (*pb.PrepayResp, error)
	QueryPayStatus(ctx context.Context, biz, outTradeNo string) (constant.PayStatus, error)
	VerifySign(r *http.Request) error
}

type PaymentLogic struct {
	payRepo repo.IPaymentRepository
	msgRepo repo.IPayMsgRepo
}

func NewPaymentLogic(repo *repo.PaymentRepository, msgRepo *repo.PayMsgRepo) *PaymentLogic {
	return &PaymentLogic{
		payRepo: repo,
		msgRepo: msgRepo,
	}
}

func (l *PaymentLogic) InitPayment(ctx context.Context, in *pb.PrepayReq) error {
	logger := logx.WithContext(ctx)
	amt, err := strconv.Atoi(in.TotalAmount)
	if err != nil {
		logger.Errorf("convert total amount to int failed, err: %v", err)
		return err
	}

	if err := l.payRepo.CreatePayment(ctx, &domain.Payment{
		Platform:   in.Platform,
		Biz:        in.Biz,
		OutTradeNo: in.OutTradeNo,
		Subject:    in.Subject,
		Amt:        int64(amt),
		Status:     constant.InitPayStatus,
		Currency:   constant.CNY,
	}); err != nil {
		logger.Errorf("create payment failed, err: %v", err)
		return err
	}

	return nil
}

func (l *PaymentLogic) UpdatePaymentStatus(ctx context.Context, biz, outTradeNo string, status constant.PayStatus) error {
	return l.payRepo.UpdatePaymentStatus(ctx, biz, outTradeNo, status)
}
