package logic

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"codexie.com/w-book-common/kafka/producer"
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
	payRepo       repo.IPaymentRepository
	msgRepo       repo.IPayMsgRepo
	kafkaProducer producer.Producer
}

func NewPaymentLogic(repo *repo.PaymentRepository, msgRepo *repo.PayMsgRepo, kafkaProducer producer.Producer) *PaymentLogic {
	return &PaymentLogic{
		payRepo:       repo,
		msgRepo:       msgRepo,
		kafkaProducer: kafkaProducer,
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

func (l *PaymentLogic) SendPayCallbackMsg(ctx context.Context, topic string, payment *domain.Payment) {
	msg, _ := json.Marshal(&PayMessage{
		OutTradeNo: payment.OutTradeNo,
		Status:     string(payment.Status),
		Amt:        int64(payment.Amt),
	})

	if err := l.kafkaProducer.SendSync(ctx, topic, string(msg), producer.WithKey(payment.OutTradeNo)); err != nil {
		//todo: 监控告警, 若短时间内发送失败多次则触发告警
		logx.Errorw("发送支付回调消息失败",
			logx.LogField{Key: "cause", Value: err.Error()},
			logx.LogField{Key: "msg", Value: msg},
		)

		// 针对少量消息发送失败使用本地消息表记录,定时任务补偿
		if err := l.msgRepo.CreateMsg(ctx, topic, string(msg)); err != nil {
			//todo: 异步重试，若失败则立即告警(消息队列支付回调消息发送失败且本地消息表记录失败)
		}
	}
}

func (l *PaymentLogic) getStatus(status string) constant.PayStatus {
	switch status {
	case TradeInit:
		return constant.InitPayStatus
	case TradeClosed:
		return constant.ClosePayStatus
	case TradeSucess, TradeFinish:
		return constant.SuceessPayStatus
	default:
		logx.Errorf("缺乏%s对应的PayStatus", status)
		return ""
	}
}
