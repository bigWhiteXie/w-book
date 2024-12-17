package limiter

import (
	"context"
	"time"
)

type LeakyBucketLimiter struct {
	interval    time.Duration //泄露频率
	tokenBucket chan struct{}
}

func NewLeakyBucketLimiter(capacity int64, interval time.Duration) *LeakyBucketLimiter {
	limiter := &LeakyBucketLimiter{
		interval:    interval,
		tokenBucket: make(chan struct{}, capacity),
	}

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C: //按照泄露频率发送请求
				limiter.tokenBucket <- struct{}{}
			}
		}
	}()
	return limiter
}

func (l *LeakyBucketLimiter) Allow(ctx context.Context) bool {
	select {
	case <-l.tokenBucket:
		return true
	case <-ctx.Done(): //等待直到请求超时
		return false
	}
}
