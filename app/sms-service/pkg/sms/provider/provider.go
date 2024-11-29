package provider

import (
	"context"
	"time"
)

const (
	Avaliable   = 1
	CrashFail   = 2
	UnAvaliable = 3
)

type SmsProvider interface {
	SendSms(ctx context.Context, phone string, args map[string]string) error

	GetName() string
	GetWeight() int
	GetStatus() int
	GetFailTime() time.Time
	GetFailCount() int

	SetFailTime(time time.Time)
	SetFailCount(count int)
	SetStatus(status int)
	SetWeight(weight int)
}
