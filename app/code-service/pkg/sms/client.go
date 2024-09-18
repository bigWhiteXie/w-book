package sms

import (
	"codexie.com/w-book-user/pkg/common/codeerr"
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

var clientMap map[string]SmsClient
var defaultClientName = "mem"

var cache *redis.Client
var maxTime int64 = 10

type SmsClient interface {
	SendSms(ctx context.Context, phone string, args map[string]string) error
}

func InitSmsClient(conf SmsConf, client *redis.Client) {
	clientMap = make(map[string]SmsClient, 2)
	clientMap["tc"] = NewTCSmsClient(conf.TC)
	clientMap["mem"] = NewMemoryClient()

	cache = client
}

func SendSms(ctx context.Context, phone string, args map[string]string) error {
	if err := VerifyCnt(ctx, phone); err != nil {
		return err
	}
	if client, ok := clientMap[defaultClientName]; ok {
		return client.SendSms(ctx, phone, args)
	}
	return nil
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

	if cnt == 1 {
		cache.Expire(ctx, key, 30*time.Minute)
	}
	return nil
}
