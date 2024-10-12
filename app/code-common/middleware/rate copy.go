package middleware

import (
	"net/http"

	"codexie.com/w-book-common/limiter"
	"github.com/zeromicro/go-zero/core/logx"
)

type BucketLimiterMiddleware struct {
	lr *limiter.TokenBucketLimiter
}

// NewCorsMiddleware 新建跨域请求处理中间件
func NewBucketLimiterMiddleware(lr *limiter.TokenBucketLimiter) *BucketLimiterMiddleware {
	return &BucketLimiterMiddleware{lr: lr}
}

func (m *BucketLimiterMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.lr.Allow(r.Context()) {
			next(w, r)
		} else {
			logx.WithContext(r.Context()).Error("rate limit")
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}
