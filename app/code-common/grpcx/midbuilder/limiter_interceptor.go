package midbuilder

import (
	"context"

	"codexie.com/w-book-common/limiter"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 针对当前进程进行限流处理
type LimiterInterceptorBuilder struct {
	limiter *limiter.SlideWindowLimiter
}

func NewLimiterInterceptorBuilder(slideLimiter *limiter.SlideWindowLimiter) *LimiterInterceptorBuilder {
	return &LimiterInterceptorBuilder{
		limiter: slideLimiter,
	}
}

func (b *LimiterInterceptorBuilder) BuildClientIntercepor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if b.limiter.Allow(ctx) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		grpc.ChainUnaryInterceptor()
		return status.Error(codes.ResourceExhausted, "请求数量达到阈值，当前请求请求被限流")
	}
}

func (b *LimiterInterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logx.Infof("fullmethod: %s", info.FullMethod)
		if b.limiter.Allow(ctx) {
			return handler(ctx, req)
		}

		return nil, status.Error(codes.ResourceExhausted, "请求数量达到阈值，当前请求请求被限流")
	}
}
