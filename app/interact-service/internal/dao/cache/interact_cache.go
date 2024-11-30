package cache

import (
	"context"

	"codexie.com/w-book-interact/internal/domain"
)

type InteractCache interface {
	//修改资源的计数信息
	IncreCntIfExist(ctx context.Context, key string, cntKind string, increment int) error

	//查询资源计数缓存
	GetStatCnt(ctx context.Context, key string) (*domain.Interaction, error)

	//缓存资源计数信息
	CacheStatCnt(ctx context.Context, key string, info *domain.Interaction) error

	//更新资源top的zset
	UpdateRedisZSet(ctx context.Context, resourceType string, fn func() ([]*domain.Interaction, error)) error

	//增加zset中资源点赞数
	IncrementLikeInZSet(ctx context.Context, resourceType string, resource int64, incre int) error

	//从zset中拿到top资源id
	GetTopFromRedisZSet(ctx context.Context, resourceType string, limit int) ([]int64, error)
}
