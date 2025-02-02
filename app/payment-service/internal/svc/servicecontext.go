package svc

import (
	"errors"

	"codexie.com/w-book-payment/internal/config"
	"codexie.com/w-book-payment/internal/logic"
	"codexie.com/w-book-payment/pkg/constant"
)

var (
	NoPlatformFound = errors.New("no payment logic found for this platform")
)

type ServiceContext struct {
	Config      config.Config
	AliPayLogic *logic.AliPayLogic
}

func NewServiceContext(c config.Config, aliPayLogic *logic.AliPayLogic) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		AliPayLogic: aliPayLogic,
	}
}

func (svc *ServiceContext) GetPayLogic(platform constant.Platform) (logic.PayLogic, error) {
	switch platform {
	case constant.ZFB:
		return svc.AliPayLogic, nil
	default:
		return nil, NoPlatformFound
	}
}
