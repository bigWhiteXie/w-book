package domain

import "codexie.com/w-book-payment/pkg/constant"

type Payment struct {
	Id int64
	// 支付平台
	Platform string
	// 支付业务
	Biz string
	// 业务单号
	OutTradeNo string
	// 支付主题
	Subject string
	// 支付金额，以分为单位
	Amt int64
	// 币种
	Currency string
	// 支付状态
	Status constant.PayStatus
}

type PaymentCallback struct {
	TradeStatus string `json:"trade_status"`
	TotalAmount string `json:"total_amount"`
	RefundFee   string `json:"refund_fee"`
	Subject     string `json:"subject"`
	OutTradeNo  string
}
