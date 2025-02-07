package svc

import (
	"errors"

	"codexie.com/w-book-account/internal/config"
)

var (
	NoPlatformFound = errors.New("no payment logic found for this platform")
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
