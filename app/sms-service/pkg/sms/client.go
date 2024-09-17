package sms

import (
	"context"
	"sync/atomic"
)

var clientMap map[string]SmsClient
var clientNames []string
var curIndex int32 = 0

type SmsClient interface {
	SendSms(ctx context.Context, phone string, args map[string]interface{}) error
}

func InitSmsClient(conf SmsConf) {
	clientMap = make(map[string]SmsClient, 2)
	clientMap["tc"] = NewTCSmsClient(conf.TC)
	clientNames = append(clientNames, "tc")
}

func SendSms(ctx context.Context, phone string, args map[string]interface{}) error {
	index := int(atomic.AddInt32(&curIndex, 1))
	name := clientNames[index%len(clientNames)]
	if client, ok := clientMap[name]; ok {
		return client.SendSms(ctx, phone, args)
	}
	return nil
}
