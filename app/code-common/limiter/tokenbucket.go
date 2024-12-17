package limiter

import (
	"context"
	"time"
)

type TokenBucketLimiter struct {
	capacity    int64         //桶容量
	interval    time.Duration //令牌速率
	tokenBucket chan struct{}
}

func NewTokenBucketLimiter(capacity int64, interval time.Duration) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		capacity:    capacity,
		interval:    interval,
		tokenBucket: make(chan struct{}, capacity),
	}

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				limiter.tokenBucket <- struct{}{}
			}
		}
	}()
	return limiter
}

func (l *TokenBucketLimiter) Allow(ctx context.Context) bool {
	select {
	case <-l.tokenBucket:
		return true
	case <-ctx.Done(): //等待直到请求超时
		return false
	}
}
