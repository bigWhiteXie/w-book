package job

import (
	"context"
	"fmt"
	"time"

	"codexie.com/w-book-common/job"
	"codexie.com/w-book-payment/internal/dao/db"
	"codexie.com/w-book-payment/internal/domain"
	"codexie.com/w-book-payment/internal/logic"
	"codexie.com/w-book-payment/internal/repo"
	"codexie.com/w-book-payment/internal/svc"
	"codexie.com/w-book-payment/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	defaultTimeExper = "*/1 * * * *"
)

type PayStatusJob struct {
	svcCtx       *svc.ServiceContext
	payRepo      repo.IPaymentRepository
	paymentLogic *logic.PaymentLogic
	timeExper    string
}

func NewPayStatusJob(svcCtx *svc.ServiceContext, repo *repo.PaymentRepository, paymentLogic *logic.PaymentLogic, opts ...job.Option) *PayStatusJob {
	payStatusJob := &PayStatusJob{
		svcCtx:       svcCtx,
		payRepo:      repo,
		paymentLogic: paymentLogic,
		timeExper:    defaultTimeExper,
	}
	for _, opt := range opts {
		opt(payStatusJob)
	}

	return payStatusJob
}

func (job *PayStatusJob) Run() error {
	size := 1000
	maxCtime := time.Now().UnixMilli()
	minCtime := maxCtime - 15*60*1000
	for {
		ctx := context.Background()
		statusChangedPayments := make([]*domain.Payment, 0)
		payments, err := job.payRepo.FindInitPaymentsByCtime(ctx, minCtime, maxCtime, size)
		if err != nil || len(payments) == 0 {
			return err
		}

		for _, payment := range payments {
			payLogic, err := job.svcCtx.GetPayLogic(constant.Platform(payment.Platform))
			if err != nil {
				continue
			}
			status, err := payLogic.QueryPayStatus(ctx, payment.Biz, payment.OutTradeNo)
			if err != nil {
				logx.Errorw("查询第三方平台支付信息异常",
					logx.LogField{Key: "platform", Value: payment.Platform},
					logx.LogField{Key: "biz", Value: payment.Biz},
					logx.LogField{Key: "outTradeNo", Value: payment.OutTradeNo},
				)
			}
			if status == constant.InitPayStatus {
				continue
			}
			payment.Status = status
			statusChangedPayments = append(statusChangedPayments, payment)
		}

		for _, pay := range statusChangedPayments {
			if err := job.payRepo.UpdatePaymentStatus(ctx, pay.Biz, pay.OutTradeNo, pay.Status); err != nil {
				if err == db.NoRowsAffectedErr {
					continue
				}

				logx.Errorw("更新支付记录状态异常",
					logx.LogField{Key: "status", Value: pay.Status},
					logx.LogField{Key: "id", Value: pay.Id},
				)
			}
			topic := fmt.Sprintf("%s-payment-callback", pay.Biz)
			job.paymentLogic.SendPayCallbackMsg(ctx, topic, &domain.Payment{
				Biz:        pay.Biz,
				OutTradeNo: pay.OutTradeNo,
				Status:     pay.Status,
				Amt:        pay.Amt,
			})
		}
		maxCtime = payments[len(payments)-1].Ctime
		if len(payments) < size {
			return nil
		}
	}
}

func (job *PayStatusJob) Name() string {
	return "payment_status_job"
}

func (job *PayStatusJob) TimeExper() string {
	return job.timeExper
}
