package logic

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"codexie.com/w-book-common/alert"
	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-common/retry"
	"codexie.com/w-book-payment/api/pb"
	"codexie.com/w-book-payment/internal/config"
	"codexie.com/w-book-payment/internal/domain"
	"codexie.com/w-book-payment/internal/types"
	"codexie.com/w-book-payment/pkg/constant"

	"github.com/smartwalle/alipay/v3"
)

type AliProductCode string

const (
	WEB_PC AliProductCode = "FAST_INSTANT_TRADE_PAY"

	TradeInit   = "WAIT_BUYER_PAY"
	TradeClosed = "TRADE_CLOSED"
	TradeSucess = "TRADE_SUCCESS"
	TradeFinish = "TRADE_FINISHED"
)

type PayMessage struct {
	OutTradeNo string `json:""`
	Status     string `json:""`
	Amt        int64  `json:""`
}

type AliPayLogic struct {
	*PaymentLogic
	client alipay.Client
	conf   config.AliPayConfig
}

func NewAliPayLogic(aliPayConf config.AliPayConfig, paymentLogic *PaymentLogic) *AliPayLogic {
	appId := aliPayConf.AppId
	privateSecret := aliPayConf.PrivateKey
	isProduction := aliPayConf.IsProduction
	var client, err = alipay.New(appId, privateSecret, isProduction)
	if err != nil {
		panic(err)
	}
	return &AliPayLogic{
		PaymentLogic: paymentLogic,
		client:       *client,
		conf:         aliPayConf,
	}
}

func (l *AliPayLogic) PrePay(ctx context.Context, in *pb.PrepayReq) (*pb.PrepayResp, error) {
	// 初始化支付记录，确保支付记录存在
	if err := l.InitPayment(ctx, in); err != nil {
		return nil, err
	}

	// 生成支付宝支付链接
	req := alipay.TradePagePay{
		Trade: alipay.Trade{
			OutTradeNo:  in.Biz + "-" + in.OutTradeNo,
			NotifyURL:   l.conf.NotifyUrl,
			TotalAmount: in.TotalAmount,
			Subject:     in.Subject,
			ProductCode: string(WEB_PC),
		},
	}

	resp, err := l.client.TradePagePay(req)
	if err != nil {
		// todo:监控该异常
		// 更新数据库失败也没关系，会有定时任务继续更新，保持最终和支付宝平台的状态一致
		l.payRepo.UpdatePaymentStatus(ctx, in.Biz, in.OutTradeNo, constant.FailPayStatus)
		return nil, codeerr.LogCodeError(ctx, strconv.Itoa(codeerr.AliPrePayErr), fmt.Sprintf("fail to invoke prepay of ali, req: %v", req))
	}

	return &pb.PrepayResp{
		PayUrl: resp.String(),
	}, nil
}

func (l *AliPayLogic) PayCallback(ctx context.Context, req *types.AliPaymentMsg) error {
	// 修改payment状态为已支付
	bizParams := strings.SplitN(req.OutTradeNo, "-", 2)
	if err := l.UpdatePaymentStatus(ctx, bizParams[0], bizParams[1], l.getStatus(req.TradeStatus)); err != nil {
		retry.AsyncRetry(time.Second*5, 3, func() error {
			return l.UpdatePaymentStatus(ctx, bizParams[0], bizParams[1], l.getStatus(req.TradeStatus))
		}, func(err error) {
			alert.Alert(alert.AlertMsg{
				Labels: map[string]string{"msg": "更新数据库中支付记录状态异常", "errcode": strconv.Itoa(codeerr.PayDBStatusErr)},
				Annotations: map[string]string{
					"cause": err.Error(),
				},
			})
		})
	}

	amt, _ := strconv.Atoi(req.TotalAmount)
	topic := fmt.Sprintf("%s-payment-callback", bizParams[0])
	l.SendPayCallbackMsg(ctx, topic, &domain.Payment{
		Biz:        bizParams[0],
		OutTradeNo: bizParams[1],
		Status:     l.getStatus(req.TradeStatus),
		Amt:        int64(amt),
	})

	return nil
}

func (l *AliPayLogic) VerifySign(r *http.Request) error {
	if err := l.client.VerifySign(r.Form); err != nil {
		return codeerr.LogCodeError(r.Context(), strconv.Itoa(codeerr.AliPaySignErr), fmt.Sprintf("fail to verify sign of ali, req: %v", r.Form))
	}
	return nil
}

func (l *AliPayLogic) GetPayment(ctx context.Context, in *pb.QueryPaymentReq) (*pb.QueryPaymentResp, error) {
	payment, err := l.payRepo.FindPaymentByBizAndOutTradeNo(ctx, in.Biz, in.OutTradeNo)
	if err != nil {
		return nil, err
	}

	return &pb.QueryPaymentResp{
		Status: string(payment.Status),
	}, nil
}

func (l *AliPayLogic) QueryPayStatus(ctx context.Context, biz, outTradeNo string) (constant.PayStatus, error) {
	resp, err := l.client.TradeQuery(ctx, alipay.TradeQuery{
		TradeNo: biz + "-" + outTradeNo,
	})
	if err != nil {
		return "", err
	}
	return l.getStatus(string(resp.TradeStatus)), nil
}
