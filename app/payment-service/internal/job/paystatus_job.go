package job

import (
	"context"

	"codexie.com/w-book-common/job"
	"codexie.com/w-book-payment/internal/repo"
	"codexie.com/w-book-payment/internal/svc"
	"codexie.com/w-book-payment/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	defaultTimeExper = "*/1 * * * *"
)

type PayStatusJob struct {
	svcCtx    *svc.ServiceContext
	payRepo   repo.IPaymentRepository
	timeExper string
}

func NewPayStatusJob(svcCtx *svc.ServiceContext, repo *repo.PaymentRepository, opts ...job.Option) *PayStatusJob {
	payStatusJob := &PayStatusJob{
		svcCtx:    svcCtx,
		payRepo:   repo,
		timeExper: defaultTimeExper,
	}
	for _, opt := range opts {
		opt(payStatusJob)
	}

	return payStatusJob
}

func (job *PayStatusJob) Run() error {
	lastId := int64(0)
	for {
		ctx := context.Background()
		statusMap := make(map[constant.PayStatus][]int64)
		payments, err := job.payRepo.FindInitPayments(ctx, lastId, 1000)
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
			if _, ok := statusMap[status]; !ok {
				statusMap[status] = make([]int64, 0)
			}
			statusMap[status] = append(statusMap[status], payment.Id)
		}

		for status, ids := range statusMap {
			if err := job.payRepo.UpdateStatusByIds(ctx, ids, status); err != nil {
				logx.Errorw("更新订单状态异常",
					logx.LogField{Key: "status", Value: status},
					logx.LogField{Key: "ids", Value: ids},
				)
			}
		}
		lastId = payments[len(payments)-1].Id
		if len(payments) < 1000 {
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
