package limiter

import (
	"context"
	"net/http"
	"sync"

	"github.com/zeromicro/go-zero/rest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CounterLimiter struct {
	limit int64 // 限制的请求数量
	count int64 // 当前请求计数
	mu    sync.RWMutex
}

// NewCounterLimiter 创建一个计数器限流器
func NewCounterLimiter(limit int64) *CounterLimiter {
	limiter := &CounterLimiter{
		limit: limit,
		count: 0,
	}

	return limiter
}

// Allow 检查是否允许通过
func (l *CounterLimiter) Allow(ctx context.Context) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.count > l.limit {
		return false
	}

	l.count++
	return true
}

func (l *CounterLimiter) Release() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.count--
}

func (l *CounterLimiter) GetHTTPLimiterMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(r.Context()) {
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
		if !l.Allow(ctx) {
			return nil, status.Errorf(codes.ResourceExhausted, "Too many requests")
		}
		return handler(ctx, req)
	}
}
