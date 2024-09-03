package limiter

import (
	"sync"
	"time"
)

type IpLimitConfig struct {
	Rate     float64 `json:"rate"`
	Capacity float64 `json:"capacity"`
}

type RateLimiter struct {
	bucket      map[string]*TokenBucket
	limitConfig IpLimitConfig
	mutex       sync.Mutex
}

type TokenBucket struct {
	rate       float64   // 速率，单位：令牌/秒
	capacity   float64   // 令牌桶容量
	tokens     float64   // 当前令牌数量
	lastUpdate time.Time // 上次更新时间
}

func NewRateLimiter(config IpLimitConfig) *RateLimiter {
	return &RateLimiter{
		bucket:      make(map[string]*TokenBucket),
		limitConfig: config,
	}
}

func (rl *RateLimiter) AllowIP(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	bucket, exists := rl.bucket[ip]
	if !exists {
		// 初始化令牌桶
		bucket = &TokenBucket{
			rate:       rl.limitConfig.Rate,     // 每秒生成10个令牌
			capacity:   rl.limitConfig.Capacity, // 令牌桶容量为10个
			tokens:     rl.limitConfig.Capacity, // 初始时令牌桶为满的状态
			lastUpdate: time.Now(),
		}
		rl.bucket[ip] = bucket
	}

	// 计算时间间隔，并根据速率生成令牌
	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	tokensToAdd := elapsed * bucket.rate

	// 更新令牌桶状态
	if tokensToAdd > 0 {
		bucket.tokens = bucket.tokens + tokensToAdd
		if bucket.tokens > bucket.capacity {
			bucket.tokens = bucket.capacity
		}
		bucket.lastUpdate = now
	}

	// 检查令牌数量是否足够
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}
