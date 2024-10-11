package sms

import (
	"context"
	"strconv"
	"time"

	"codexie.com/w-book-code/internal/kafka/producer"
	"codexie.com/w-book-code/internal/model"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-user/pkg/common/codeerr"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

var topic = "sms-topic"

var smsService *SmsService
var smsRepo repo.SmsRepo
var provider *producer.KafkaProducer
var cache *redis.Client
var maxTime int64 = 10

// mockgen -source=pkg/sms/client.go -destination mocks/pkg/sms/sms_mock.go
type SmsClient interface {
	SendSms(ctx context.Context, phone string, args map[string]string) error
}

func InitSmsClient(conf SmsConf, client *redis.Client, repo repo.SmsRepo, kafkaProvider *producer.KafkaProducer) {
	smsService = NewSmsService(conf.TC, conf.Memory)
	smsRepo = repo
	provider = kafkaProvider
	cache = client
}

func SendSms(ctx context.Context, phone string, args map[string]string) error {
	if err := VerifyCnt(ctx, phone); err != nil {
		return err
	}
	//短信记录落库
	record := model.NewSmsRecord(phone, args)
	if err := smsRepo.SaveSmsRecord(ctx, record); err != nil {
		logx.Errorf("[send code] %v", err)
		return err
	}
	//同步or异步
	if IsSyncSend() {
		if err := smsService.SendSms(ctx, phone, args); err != nil {
			logx.Errorf("[SendSms] 同步发送验证码失败,cause:%v", err)
			provider.Send(ctx, topic, strconv.Itoa(record.ID), func(err error) {
				logx.Errorf("fail to send msg to kafka, cause:%s", err)
			})
		}
		return nil
	}
	// 响应时间过长，异步发送短信
	provider.Send(ctx, topic, strconv.Itoa(record.ID), func(err error) {
		logx.Errorf("fail to send msg to kafka, cause:%s", err)
	})

	return nil
}

func MustSendSms(ctx context.Context, phone string, args map[string]string) error {
	return smsService.MustSendSms(ctx, phone, args)
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

func IsSyncSend() bool {
	if len(smsService.RespTimeQueue) == 0 {
		return true
	}

	var total time.Duration
	for _, d := range smsService.RespTimeQueue {
		total += d
	}
	if avgRespTime := total / time.Duration(len(smsService.RespTimeQueue)); avgRespTime <= 1*time.Second {
		return true
	}

	return false
}
