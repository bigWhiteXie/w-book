package sms

import (
	"context"
	"math/rand"
	"time"
)

type SmsProvider struct {
	Name      string //服务商
	Weight    int    // 权重
	Status    int    // 0 不可用 | 1 可用 | 2 暂时可用
	client    SmsClient
	FailCount int
	FailLimit int
}

type SmsService struct {
	smsProviders map[string]*SmsProvider
}

func NewSmsService(tcConfig Tencent, memoryConf Memeory) *SmsService {
	providers := make(map[string]*SmsProvider)
	tcClient := NewTCSmsClient(tcConfig)
	memClient := NewMemoryClient()

	providers[tcConfig.Name] = &SmsProvider{
		Name:      tcConfig.Name,
		Weight:    tcConfig.Weight,
		Status:    1,
		client:    tcClient,
		FailCount: 0,
		FailLimit: 3,
	}
	providers[memoryConf.Name] = &SmsProvider{
		Name:      memoryConf.Name,
		Weight:    memoryConf.Weight,
		Status:    1,
		client:    memClient,
		FailCount: 0,
		FailLimit: 3,
	}

	return &SmsService{
		smsProviders: providers,
	}
}

func (s *SmsService) SendSms(ctx context.Context, phone string, args map[string]string) error {
	provider := s.selectProvider()
	startTime := time.Now()
	err := provider.client.SendSms(ctx, phone, args)
	responseTime := time.Since(startTime)
	if err != nil || responseTime >= 3*time.Second {
		s.markProviderUnavailable(provider, 60*time.Second, err != nil)
		if err != nil {
			// TODO 添加到数据库中，等待补偿任务
		}
	}

	return nil
}
func (s *SmsService) selectProvider() *SmsProvider {
	totalWeight := 0
	for _, p := range s.smsProviders {
		if p.Status == 1 {
			totalWeight += p.Weight
		}
	}

	randNum := rand.Intn(totalWeight)
	for _, p := range s.smsProviders {
		if p.Status == 1 {
			randNum -= p.Weight
			if randNum < 0 {
				return p
			}
		}
	}
	return nil
}

// 当发送失败时调用该方法，会自动将连续失败N次的供应商标记为失败，并在cooldownTime后重新标记为可用
func (s *SmsService) markProviderUnavailable(provider *SmsProvider, cooldownTime time.Duration, fastFail bool) {
	provider.FailCount++
	if fastFail || provider.FailCount >= provider.FailLimit || provider.Status == 2 {
		provider.Status = 0
		go func(p *SmsProvider) {
			time.Sleep(cooldownTime)
			p.Status = 2
			provider.FailCount = 0
		}(provider)
	}
}
