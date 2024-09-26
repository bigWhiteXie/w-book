package sms

import (
	"context"
	"time"

	"codexie.com/w-book-user/pkg/common/codeerr"
	"github.com/redis/go-redis/v9"
)

var smsService *SmsService

var cache *redis.Client
var maxTime int64 = 10

// mockgen -source=pkg/sms/client.go -destination mocks/pkg/sms/sms_mock.go
type SmsClient interface {
	SendSms(ctx context.Context, phone string, args map[string]string) error
}

func InitSmsClient(conf SmsConf, client *redis.Client) {
	smsService = NewSmsService(conf.TC, conf.Memory)
	cache = client
}

func SendSms(ctx context.Context, phone string, args map[string]string) error {
	if err := VerifyCnt(ctx, phone); err != nil {
		return err
	}

	return smsService.SendSms(ctx, phone, args)
}

func VerifyCnt(ctx context.Context, phone string) error {
	key := "sms:cnt:" + phone
	cnt, err := cache.IncrBy(ctx, key, 1).Result()
	if err != nil {
		return err
	}
	if cnt > maxTime {
		return codeerr.WithCode(codeerr.SmsFrequentERR, "%s send sms arrived %d", phone, cnt)
	}

	// 首次发送短信设置过期时间
	if cnt == 1 {
		cache.Expire(ctx, key, 30*time.Minute)
	}
	return nil
}
