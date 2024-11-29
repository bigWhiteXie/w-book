package provider

import (
	"context"
	"time"

	"codexie.com/w-book-code/pkg/sms"
	"github.com/zeromicro/go-zero/core/logx"
)

type MemorySmsProvider struct {
	status       int
	weight       int
	failCount    int
	lastFailTime time.Time
}

func NewMemoryClient(c sms.Memeory) SmsProvider {
	return &MemorySmsProvider{
		weight: c.Weight,
		status: Avaliable,
	}
}

func (client *MemorySmsProvider) SendSms(ctx context.Context, phone string, args map[string]string) error {
	logx.WithContext(ctx).Infof("向[%s]发送短信：%v", phone, args)
	return nil
}

func (client *MemorySmsProvider) GetName() string {
	return "memory"
}

func (client *MemorySmsProvider) GetWeight() int {
	return client.weight
}

func (client *MemorySmsProvider) GetStatus() int {
	return client.status
}

func (client *MemorySmsProvider) GetFailCount() int {
	return client.failCount
}

func (client *MemorySmsProvider) GetFailTime() time.Time {
	return client.lastFailTime
}

func (client *MemorySmsProvider) SetFailTime(t time.Time) {
	client.lastFailTime = t
}

func (client *MemorySmsProvider) SetFailCount(count int) {
	client.failCount = count
}

func (client *MemorySmsProvider) SetWeight(weight int) {
	client.weight = weight
}

func (client *MemorySmsProvider) SetStatus(status int) {
	client.status = status
}
