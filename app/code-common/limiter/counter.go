package limiter

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/rest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CounterLimiter struct {
	limit    int           // 限制的请求数量
	interval time.Duration // 时间窗口
	count    int64         // 当前请求计数
	lastTime time.Time
	mu       sync.RWMutex
}

// NewCounterLimiter 创建一个计数器限流器
func NewCounterLimiter(limit int, interval time.Duration) *CounterLimiter {
	limiter := &CounterLimiter{
		limit:    limit,
		interval: interval,
		count:    0,
		lastTime: time.Now(),
	}

	return limiter
}

// Allow 检查是否允许通过
func (l *CounterLimiter) Allow() bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Sub(l.lastTime) > l.interval {
		l.lastTime = now
		l.count += 1
	}
	if l.count += 1; l.count > int64(l.limit) {
		return false
	}

	return true
}

func (l *CounterLimiter) GetHTTPLimiterMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow() {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Too many requests"))
				return
			}
			next(w, r)
		}
	}
}

func (l *CounterLimiter) GetGrpcLimiterInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !l.Allow() {
			return nil, status.Errorf(codes.ResourceExhausted, "Too many requests")
		}
		return handler(ctx, req)
	}
}
