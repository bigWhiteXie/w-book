package cache

import (
	"context"

	"codexie.com/w-book-interact/internal/domain"
)

type InteractCache interface {
	//修改资源的计数信息
	IncreCntIfExist(ctx context.Context, key string, cntKind string, increment int) error

	//查询资源计数缓存
	GetStatCnt(ctx context.Context, key string) (*domain.StatCnt, error)

	//缓存资源计数信息
	CacheStatCnt(ctx context.Context, key string, info *domain.StatCnt) error
}
