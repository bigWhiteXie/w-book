package logic

import (
	"context"
	"time"

	"codexie.com/w-book-common/metric"
)

type PrometheusSmsLogic struct {
	SmsService SmsService
}

func NewPrometheusSmsLogic(labelConf metric.ConstMetricLabelsConf, smsService SmsService) *PrometheusSmsLogic {
	metric.InitSmsMetric(labelConf)
	return &PrometheusSmsLogic{SmsService: smsService}
}

func (s *PrometheusSmsLogic) SendSms(ctx context.Context, phone string, args map[string]string) (string, error) {
	metric.SmsConcurrentGauge.Add(1)
	begin := time.Now()
	brand, err := s.SmsService.SendSms(ctx, phone, args)
	metric.SmsSendCountCounter.WithLabelValues(brand).Inc()
	metric.SmsConcurrentGauge.Sub(1)
	metric.SmsSendLatencyHistogram.WithLabelValues(brand).Observe(float64(time.Since(begin).Milliseconds()))
	if err != nil {
		metric.SmsSendErrCounter.WithLabelValues(brand, err.Error()).Inc()
	}
	return brand, err
}
