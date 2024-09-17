package middleware

import (
	"codexie.com/w-book-user/pkg/limiter"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"strings"
)

type IpLimiterMiddleware struct {
	lr *limiter.RateLimiter
}

// NewCorsMiddleware 新建跨域请求处理中间件
func NewLimiterMiddleware(lr *limiter.RateLimiter) *IpLimiterMiddleware {
	return &IpLimiterMiddleware{lr: lr}
}

func (m *IpLimiterMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if m.lr.AllowIP(ip) {
			next(w, r)
		} else {
			logx.WithContext(r.Context()).Error("rate limit")
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}

func getClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 头中获取 IP 地址
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For 头可能包含多个 IP 地址，用逗号分隔，取第一个
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// 尝试从 X-Real-Ip 头中获取 IP 地址
	realIP := r.Header.Get("X-Real-Ip")
	if realIP != "" {
		return realIP
	}

	// 最后从 RemoteAddr 中获取 IP 地址
	ip := r.RemoteAddr
	// RemoteAddr 可能包含端口号，需要去掉
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}
