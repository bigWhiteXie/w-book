package logic

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"codexie.com/w-book-common/kafka/producer"
	"codexie.com/w-book-payment/api/pb"
	"codexie.com/w-book-payment/internal/config"
	"codexie.com/w-book-payment/internal/types"
	"codexie.com/w-book-payment/pkg/constant"

	"github.com/smartwalle/alipay/v3"
	"github.com/zeromicro/go-zero/core/logx"
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
}

type AliPayLogic struct {
	*PaymentLogic
	client        alipay.Client
	kafkaProducer producer.Producer
}

func NewAliPayLogic(aliPayConf config.AliPayConfig, paymentLogic *PaymentLogic, p producer.Producer) *AliPayLogic {
	appId := aliPayConf.AppId
	privateSecret := aliPayConf.PrivateKey
	isProduction := aliPayConf.IsProduction
	var client, err = alipay.New(appId, privateSecret, isProduction)
	if err != nil {
		panic(err)
	}
	return &AliPayLogic{
		PaymentLogic:  paymentLogic,
		client:        *client,
		kafkaProducer: p,
	}
}

func (l *AliPayLogic) PayCallback(ctx context.Context, req *types.AliPaymentMsg) error {
	// 修改payment状态为已支付
	bizParams := strings.SplitN(req.OutTradeNo, "-", 2)
	if err := l.payRepo.UpdatePaymentStatus(ctx, bizParams[0], bizParams[1], l.getStatus(req.TradeStatus)); err != nil {
		//todo: 此时用户已经付完钱了，但是修改状态失败，应该重试，若重试仍然不行则告警
	}
	topic := bizParams[0]
	//在本地消息表中创建记录，并发送消息到消息队列，保证消息一定发出
	msg, _ := json.Marshal(&PayMessage{
		OutTradeNo: req.OutTradeNo,
		Status:     string(l.getStatus(req.TradeStatus)),
	})
	if err := l.kafkaProducer.SendSync(ctx, bizParams[0], string(msg)); err != nil {
		//todo: 监控告警, 若短时间内发送失败多次则触发告警
		logx.Errorw("发送支付回调消息失败",
			logx.LogField{Key: "cause", Value: err.Error()},
			logx.LogField{Key: "msg", Value: msg},
		)

		// 针对少量消息发送失败使用本地消息表记录
		if err := l.msgRepo.CreateMsg(ctx, topic, string(msg)); err != nil {
			//todo: 异步重试，若失败则立即告警(消息队列支付回调消息发送失败且本地消息表记录失败)
		}
	}

	return nil
}

func (l *AliPayLogic) VerifySign(r *http.Request) error {
	return l.client.VerifySign(r.Form)
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

func (l *AliPayLogic) PrePay(ctx context.Context, in *pb.PrepayReq) (*pb.PrepayResp, error) {
	logger := logx.WithContext(ctx)
	// 支付记录落库
	if err := l.InitPayment(ctx, in); err != nil {
		return nil, err
	}

	// 生成支付宝支付链接
	req := alipay.TradePagePay{
		Trade: alipay.Trade{
			OutTradeNo: in.Biz + "-" + in.OutTradeNo,
			// NotifyURL: "http://127.0.0.1",
			TotalAmount: in.TotalAmount,
			Subject:     in.Subject,
			ProductCode: string(WEB_PC),
		},
	}

	resp, err := l.client.TradePagePay(req)
	if err != nil {
		logger.Errorf("fail to invoke prepay of ali, err: %v", err)
		// 更新数据库失败也没关系，会有定时任务继续更新，保持最终和支付宝平台的状态一致
		l.payRepo.UpdatePaymentStatus(ctx, in.Biz, in.OutTradeNo, constant.FailPayStatus)
		return nil, err
	}
	return &pb.PrepayResp{
		PayUrl: resp.String(),
	}, nil
}

func (l *AliPayLogic) getStatus(status string) constant.PayStatus {
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
