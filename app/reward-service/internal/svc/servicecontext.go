package svc

import (
	"codexie.com/w-book-reward/internal/config"

	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config config.Config
	Cache  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
