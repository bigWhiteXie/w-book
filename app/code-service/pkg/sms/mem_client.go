package sms

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type MemoryClient struct {
}

func NewMemoryClient() *MemoryClient {
	return &MemoryClient{}
}

func (client *MemoryClient) SendSms(ctx context.Context, phone string, args map[string]string) error {
	logx.WithContext(ctx).Infof("向[%s]发送短信：%v", phone, args)
	return nil
}
