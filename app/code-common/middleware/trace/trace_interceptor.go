package trace

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type TraceConfig struct {
	Addr string `json:""`
	Name string `json:""`
}

type TraceInterceptorBuilder struct {
	name       string
	tp         trace.Tracer
	propagator propagation.TextMapPropagator
}

func NewTraceInterceptorBuilder(config TraceConfig) *TraceInterceptorBuilder {
	tracer := otel.GetTracerProvider().Tracer(config.Name)
	return &TraceInterceptorBuilder{
		name:       config.Name,
		tp:         tracer,
		propagator: otel.GetTextMapPropagator(),
	}
}

// GrpcClientInterceptor 返回 gRPC 客户端拦截器
func (b *TraceInterceptorBuilder) GrpcClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		tr := b.tp
		peerAddr := PeerFromCtx(ctx)
		methodName, attr := SpanInfo(method, peerAddr)
		ctx, span := tr.Start(ctx, methodName, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attr...))
		defer span.End()
		// 把该span注入到md中，最终注入到ctx中传递给服务端
		b.propagator.Inject(ctx, &metadataSupplier{
			metadata: &md,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			s, ok := status.FromError(err)
			if ok {
				span.SetStatus(codes.Error, s.Message())
			} else {
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		}
		return nil
	}
}

// GrpcServerInterceptor 返回 gRPC 服务端拦截器
// 从metadata中解析trace上下文并注入ctx中，此外需要处理异常
func (b *TraceInterceptorBuilder) GrpcServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		bag, spanCtx := Extract(ctx, b.propagator)
		ctx = baggage.ContextWithBaggage(ctx, bag)
		peerAddr := PeerFromCtx(ctx)
		methodName, attr := SpanInfo(info.FullMethod, peerAddr)
		ctx, span := b.tp.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), methodName, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attr...))
		defer span.End()
		resp, err = handler(ctx, req)
		if err != nil {
			s, ok := status.FromError(err)
			if ok {
				span.SetStatus(codes.Error, s.Message())

			} else {
				span.SetStatus(codes.Error, err.Error())
			}
			return nil, err
		}

		return resp, nil
	}
}

// HttpMiddleware 返回 HTTP 中间件
func (b *TraceInterceptorBuilder) HttpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r)
	})
}

func PeerFromCtx(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p == nil {
		return ""
	}

	return p.Addr.String()
}
