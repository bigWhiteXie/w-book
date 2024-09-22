package sms

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"codexie.com/w-book-user/pkg/common/codeerr"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	SmsFailErr = errors.New("短信发送失败")
)

type SmsProvider struct {
	Name      string //服务商
	Weight    int    // 权重
	Status    int    // 0 不可用 | 1 可用 | 2 暂时可用
	Client    SmsClient
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
		Client:    tcClient,
		FailCount: 0,
		FailLimit: 3,
	}
	providers[memoryConf.Name] = &SmsProvider{
		Name:      memoryConf.Name,
		Weight:    memoryConf.Weight,
		Status:    1,
		Client:    memClient,
		FailCount: 0,
		FailLimit: 3,
	}

	return &SmsService{
		smsProviders: providers,
	}
}

func (s *SmsService) SendSms(ctx context.Context, phone string, args map[string]string) error {
	for {
		provider := s.selectProvider()
		if provider == nil {
			return codeerr.WithCode(codeerr.SmsNotAvaliableErr, "sms provider is nil")
		}
		startTime := time.Now()
		err := provider.Client.SendSms(ctx, phone, args)
		responseTime := time.Since(startTime)
		if err != nil || responseTime >= 3*time.Second {
			s.markProviderUnavailable(provider, 60*time.Second, err != nil)
			if err != nil {
				logx.Errorf("[%s] 短信发送失败,cause:%v", provider.Name, err)
			} else {
				logx.Errorf("[%s]短信发送失败,响应时间超过3秒,请检查短信服务商", provider.Name)
			}
			continue
		}
		// 重置provider状态
		provider.Status = 1
		provider.FailCount = 0
		logx.Infof("短信发送成功，耗时：%v", responseTime)
		return nil
	}
}
func (s *SmsService) selectProvider() *SmsProvider {
	totalWeight := 0
	for _, p := range s.smsProviders {
		if p.Status == 1 {
			totalWeight += p.Weight
		}
	}
	// 此时没有可用的服务商
	if totalWeight == 0 {
		return nil
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
