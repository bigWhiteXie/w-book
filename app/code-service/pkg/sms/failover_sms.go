package sms

import (
	"context"
	"errors"
	"math/rand"
	"sync/atomic"
	"time"

	"codexie.com/w-book-common/common/codeerr"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	SmsFailErr = errors.New("短信发送失败")
)

var count uint64 = 0

type SmsProvider struct {
	Name         string //服务商
	Weight       int    // 权重
	Status       int    // 0 不可用 | 1 可用 | 2 暂时可用
	Client       SmsClient
	FailCount    int
	FailLimit    int
	LastFailTime time.Time
}

type SmsService struct {
	smsProviders  map[string]*SmsProvider
	RespTimeQueue [10]time.Duration
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
	provider := s.selectProvider()
	if provider == nil {
		return codeerr.WithCode(codeerr.SmsNotAvaliableErr, "sms provider is nil")
	}

	startTime := time.Now()
	err := provider.Client.SendSms(ctx, phone, args)
	responseTime := time.Since(startTime)
	idx := (atomic.AddUint64(&count, 1)) % 10
	s.RespTimeQueue[int(idx)] = responseTime
	if err != nil || responseTime >= 3*time.Second {
		s.markProviderUnavailable(provider, 60*time.Second, err != nil)
		if err != nil {
			logx.Errorf("[%s] 短信发送失败,cause:%v", provider.Name, err)
			return err
		} else {
			logx.Errorf("[%s]短信发送失败,响应时间超过3秒,请检查短信服务商", provider.Name)
			return nil
		}
	}
	// 重置provider状态
	provider.Status = 1
	provider.FailCount = 0
	logx.Infof("短信发送成功，耗时：%v", responseTime)
	return nil
}

func (s *SmsService) MustSendSms(ctx context.Context, phone string, args map[string]string) error {
	for {
		codeErr := codeerr.WithCodeErr{}
		if err := s.SendSms(ctx, phone, args); errors.As(err, codeErr) && codeErr.Code == codeerr.SmsNotAvaliableErr {
			return &codeErr
		}
	}
}
func (s *SmsService) selectProvider() *SmsProvider {
	totalWeight := 0
	for _, p := range s.smsProviders {
		if p.Status != 0 {
			totalWeight += p.Weight
		} else if time.Since(p.LastFailTime) >= 60*time.Second {
			p.Status = 2
			totalWeight += p.Weight
		}
	}
	// 此时没有可用的服务商
	if totalWeight == 0 {
		return nil
	}

	randNum := rand.Intn(totalWeight)
	for _, p := range s.smsProviders {
		if p.Status != 0 {
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
		provider.LastFailTime = time.Now()
	}
}
