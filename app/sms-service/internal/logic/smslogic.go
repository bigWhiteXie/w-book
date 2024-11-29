package logic

import (
	"context"
	"math/rand"
	"time"

	"codexie.com/w-book-code/pkg/sms/provider"
	"codexie.com/w-book-common/common/codeerr"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	failLimit = 3
)

type SmsService interface {
	SendSms(ctx context.Context, phone string, args map[string]string) error
}

type ProviderSmsLogic struct {
	smsProviders map[string]provider.SmsProvider
	failLimit    int
}

func NewProviderSmsLogic(providers ...provider.SmsProvider) *ProviderSmsLogic {
	smsProviders := make(map[string]provider.SmsProvider, len(providers))
	for _, provider := range providers {
		smsProviders[provider.GetName()] = provider
	}
	return &ProviderSmsLogic{
		smsProviders: smsProviders,
		failLimit:    failLimit,
	}
}

func (s *ProviderSmsLogic) SendSms(ctx context.Context, phone string, args map[string]string) error {
	p := s.selectProvider()
	if p == nil {
		return codeerr.WithCode(codeerr.SmsNotAvaliableErr, "no avaliable sms provider")
	}

	startTime := time.Now()
	err := p.SendSms(ctx, phone, args)
	responseTime := time.Since(startTime)

	if err != nil || responseTime >= 3*time.Second {
		s.markProviderUnavailable(p, 60*time.Second, err != nil)
		if err != nil {
			return err
		} else {
			logx.Errorf("[%s]响应时间超过3秒,请检查短信服务商是否故障", p.GetName())
			return nil
		}
	}

	// 重置provider状态
	p.SetStatus(provider.Avaliable)

	logx.Infof("短信发送成功，耗时：%v", responseTime)
	return nil
}

func (s *ProviderSmsLogic) selectProvider() provider.SmsProvider {
	totalWeight := 0
	for _, p := range s.smsProviders {
		if p.GetStatus() != provider.UnAvaliable {
			totalWeight += p.GetWeight()
		} else if time.Since(p.GetFailTime()) >= 60*time.Second {
			p.SetStatus(provider.CrashFail)
			totalWeight += p.GetWeight()
		}
	}
	// 此时没有可用的服务商
	if totalWeight == 0 {
		return nil
	}

	randNum := rand.Intn(totalWeight)
	for _, p := range s.smsProviders {
		if p.GetStatus() != provider.UnAvaliable {
			randNum -= p.GetWeight()
			if randNum < 0 {
				return p
			}
		}
	}
	return nil
}

// 当发送失败时调用该方法，会自动将连续失败N次的供应商标记为失败，并在cooldownTime后重新标记为可用
func (s *ProviderSmsLogic) markProviderUnavailable(p provider.SmsProvider, cooldownTime time.Duration, fastFail bool) {
	p.SetFailCount(p.GetFailCount() + 1)
	if fastFail || p.GetFailCount() >= s.failLimit || p.GetStatus() == provider.CrashFail {
		p.SetStatus(provider.UnAvaliable)
		p.SetFailTime(time.Now())
	}
}
