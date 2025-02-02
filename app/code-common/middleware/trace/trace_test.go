package trace

import (
	"context"
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	status "google.golang.org/grpc/status"
)

var serverTraceBuilder *TraceInterceptorBuilder
var clientTraceBuilder *TraceInterceptorBuilder

func TestTraceInteceptor(t *testing.T) {
	suite.Run(t, &TraceInterceptorSuite{})
}

type TraceInterceptorSuite struct {
	suite.Suite
	tp *sdktrace.TracerProvider
}

func (s *TraceInterceptorSuite) SetupSuite() {
	grpcOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:14317"),
	}
	exp, err := otlptracegrpc.New(context.Background(), grpcOpts...)
	if err != nil {
		panic(err)
	}
	kv := semconv.ServiceNameKey.String("test-trace")
	// traceProvider可是指定采样率、exporter以及每个span都会携带的resource信息
	opts := []sdktrace.TracerProviderOption{
		// Set the sampling rate based on the parent span to 100%
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1.0))),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewSchemaless(kv)),
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	s.tp = sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(s.tp)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logx.Errorf("[otel] error: %v", err)
	}))
	serverTraceBuilder = NewTraceInterceptorBuilder(TraceConfig{
		Name: "test-server",
	})

	clientTraceBuilder = NewTraceInterceptorBuilder(TraceConfig{
		Name: "test-client",
	})

}

func (s *TraceInterceptorSuite) TearDownTest() {
	s.tp.Shutdown(context.Background())
}

func (s *TraceInterceptorSuite) TestTrace() {
	// 创建grpc server
	serverAddr := "127.0.0.1:24423"
	go func() {
		startTraceServer(0, serverAddr)
	}()

	// 创建grpc客户端
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(clientTraceBuilder.GrpcClientInterceptor()),
	)
	if err != nil {
		panic(err)
	}
	client := NewTraceServerClient(conn)
	client.GetUser(context.Background(), &GetUserReq{
		Uid: "11",
	})
	time.Sleep(5 * time.Second)
	log.Println("test finish")
}

type traceServer struct {
	UnimplementedTraceServerServer
	index int
}

func (s *traceServer) GetUser(cex context.Context, req *GetUserReq) (*GetUserResp, error) {
	// log.Printf("服务实例:%d, 收到请求: %v\n", s.index, req)
	if s.index == 2 {
		//返回grpc的unavaliable异常
		return nil, status.Errorf(codes.Unavailable, "service is temporarily unavailable")
	}
	return &GetUserResp{
		Uid:  strconv.Itoa(s.index),
		Name: "test",
	}, nil
}

func startTraceServer(index int, addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(serverTraceBuilder.GrpcServerInterceptor()),
	)

	RegisterTraceServerServer(s, &traceServer{index: index})
	// 创建健康检查服务
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	log.Printf("服务端正在监听: %s\n", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务端启动失败: %v", err)
	}
}
