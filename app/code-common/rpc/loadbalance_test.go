package rpc

import (
	"context"
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-common/grpcx/balancer"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	status "google.golang.org/grpc/status"
)

func TestDefaultLoadBalance(t *testing.T) {
	// 1. 连接 etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2479"}, // etcd 地址
		DialTimeout: 5 * time.Second,            // 连接超时时间
	})
	if err != nil {
		t.Fatalf("连接 etcd 失败: %v", err)
	}
	defer cli.Close()

	serviceKey := "code/rpc" // 服务名称

	// 2. 服务注册
	em, err := endpoints.NewManager(cli, serviceKey)
	if err != nil {
		t.Fatalf("创建 endpoint.Manager 失败: %v", err)
	}

	// 注册服务实例
	serviceEndpoints := []string{"127.0.0.1:24423", "127.0.0.1:24424", "127.0.0.1:24425"}
	lease, err := cli.Grant(context.Background(), 10) // 10 秒的租约
	for i, addr := range serviceEndpoints {
		go startServer(i, addr)

		if err != nil {
			t.Fatalf("创建租约失败: %v", err)
		}
		endpoint := endpoints.Endpoint{
			Addr: addr,
			Metadata: map[string]string{
				"weight": "10",
			},
		}

		err = em.AddEndpoint(context.Background(), serviceKey+"/"+addr, endpoint, clientv3.WithLease(lease.ID))
		if err != nil {
			t.Fatalf("服务发布失败: %v", err)
		}
		log.Printf("服务发布成功: %s -> %s\n", serviceKey, addr)
	}

	// 3. 实现 gRPC Resolver
	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		t.Fatalf("创建 etcd resolver 失败: %v", err)
	}

	// 3.1 查询服务实例列表
	kvMap, _ := em.List(context.Background())
	for k, v := range kvMap {
		log.Printf("服务实例：%s->%v\n", k, v)
	}

	// 4. 创建 grpc 连接
	conn, err := grpc.NewClient(
		"etcd:///"+serviceKey, // 使用 etcd resolver
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("连接 gRPC 服务失败: %v", err)
	}
	log.Printf("grpc conn target:%s\n", conn.Target())
	log.Println("gRPC 连接成功，服务发现完成")

	// 5. 构建 grpc 客户端
	codeClient := pb.NewCodeClient(conn)

	// 6. 模拟调用 RPC 方法
	for i := 0; i < 10; i++ {
		_, err := codeClient.SendCode(context.Background(), &pb.SendCodeReq{
			Phone: "12345677",
		})
		if err != nil {
			log.Printf("调用 RPC 方法失败: %v\n", err)
		} else {
			log.Printf("调用 RPC 方法成功\n")
		}
		time.Sleep(500 * time.Millisecond) // 模拟间隔
	}
	time.Sleep(5 * time.Second)

	//开始服务注销
	cli.Revoke(context.Background(), lease.ID)
	conn.Close()
}

func TestWeightRoubin(t *testing.T) {
	// 1. 连接 etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2479"}, // etcd 地址
		DialTimeout: 5 * time.Second,            // 连接超时时间
	})
	if err != nil {
		t.Fatalf("连接 etcd 失败: %v", err)
	}
	defer cli.Close()

	serviceKey := "user/rpc" // 服务名称

	// 2. 服务注册
	em, err := endpoints.NewManager(cli, serviceKey)
	if err != nil {
		t.Fatalf("创建 endpoint.Manager 失败: %v", err)
	}

	// 注册服务实例
	serviceEndpoints := []string{"127.0.0.1:24423", "127.0.0.1:24424", "127.0.0.1:24425"}
	lease, err := cli.Grant(context.Background(), 10) // 10 秒的租约

	// 一定要开启续约，否则服务地址全部注销client无法访问
	ctx, cancel := context.WithCancel(context.Background())
	go cli.KeepAlive(ctx, lease.ID)
	for i, addr := range serviceEndpoints {
		go startServer(i, addr)

		if err != nil {
			t.Fatalf("创建租约失败: %v", err)
		}
		endpoint := endpoints.Endpoint{
			Addr: addr,
			Metadata: map[string]string{
				"weight": strconv.Itoa(10 * (i + 1)),
			},
		}

		err = em.AddEndpoint(context.Background(), serviceKey+"/"+addr, endpoint, clientv3.WithLease(lease.ID))
		if err != nil {
			t.Fatalf("服务发布失败: %v", err)
		}
		log.Printf("服务发布成功: %s -> %s\n", serviceKey, addr)
	}

	// 3. 实现 gRPC Resolver
	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		t.Fatalf("创建 etcd resolver 失败: %v", err)
	}

	// 3.1 查询服务实例列表
	kvMap, _ := em.List(context.Background())
	for k, v := range kvMap {
		log.Printf("服务实例：%s->%v\n", k, v)
	}
	balancer.InitWeightRoubin()
	// 4. 创建 grpc 连接
	serviceConfig := `{
		"loadBalancingConfig": [{"weighted_round_robin": {}}],
		"methodConfig": [
			{
				"name": [
					{"service": "user.UserService", "method": "GetUser"}
				],
				"retryPolicy": {
					"maxAttempts": 3,
					"initialBackoff": "0.1s",
					"maxBackoff": "1s",
					"backoffMultiplier": 2.0,
					"retryableStatusCodes": ["UNAVAILABLE","RESOURCE_EXHAUSTED"]
				}
			}
		]
	}`
	conn, err := grpc.Dial(
		"etcd:///"+serviceKey, // 使用 etcd resolver
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("连接 gRPC 服务失败: %v", err)
	}
	log.Printf("grpc conn target:%s\n", conn.Target())
	log.Println("gRPC 连接成功，服务发现完成")

	// 5. 构建 grpc 客户端
	userClient := NewUserServiceClient(conn)

	// 6. 模拟调用 RPC 方法
	log.Println("==================开始发送grpc请求==================")
	for i := 0; i < 10; i++ {
		_, err := userClient.GetUser(context.Background(), &GetUserReq{
			Uid: "12345677",
		})
		if err != nil {
			log.Printf("调用 RPC 方法失败: %v\n", err)
		} else {
			log.Printf("调用 RPC 方法成功\n")
		}
		// time.Sleep(1 * time.Second)
	}
	//等待所有请求完成再注销服务
	cancel()
	time.Sleep(5 * time.Second)
	cli.Revoke(context.Background(), lease.ID)
	conn.Close()
}

func TestConsistentBalance(t *testing.T) {
	// 1. 连接 etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2479"}, // etcd 地址
		DialTimeout: 5 * time.Second,            // 连接超时时间
	})
	if err != nil {
		t.Fatalf("连接 etcd 失败: %v", err)
	}
	defer cli.Close()

	serviceKey := "user/rpc" // 服务名称

	// 2. 服务注册
	em, err := endpoints.NewManager(cli, serviceKey)
	if err != nil {
		t.Fatalf("创建 endpoint.Manager 失败: %v", err)
	}

	// 注册服务实例
	serviceEndpoints := []string{"127.0.0.1:24423", "127.0.0.1:24424", "127.0.0.1:24425"}
	lease, err := cli.Grant(context.Background(), 10) // 10 秒的租约

	// 一定要开启续约，否则服务地址全部注销client无法访问
	ctx, cancel := context.WithCancel(context.Background())
	go cli.KeepAlive(ctx, lease.ID)
	for i, addr := range serviceEndpoints {
		go startServer(i, addr)

		if err != nil {
			t.Fatalf("创建租约失败: %v", err)
		}
		endpoint := endpoints.Endpoint{
			Addr: addr,
			Metadata: map[string]string{
				"weight": strconv.Itoa(10 * (i + 1)),
			},
		}

		err = em.AddEndpoint(context.Background(), serviceKey+"/"+addr, endpoint, clientv3.WithLease(lease.ID))
		if err != nil {
			t.Fatalf("服务发布失败: %v", err)
		}
		log.Printf("服务发布成功: %s -> %s\n", serviceKey, addr)
	}

	// 3. 实现 gRPC Resolver
	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		t.Fatalf("创建 etcd resolver 失败: %v", err)
	}

	// 3.1 查询服务实例列表
	kvMap, _ := em.List(context.Background())
	for k, v := range kvMap {
		log.Printf("服务实例：%s->%v\n", k, v)
	}
	balancer.InitConsistentBalancer("uid")
	// 4. 创建 grpc 连接
	serviceConfig := `{
		"loadBalancingConfig": [{"consistance_hash_balance": {}}],
		"methodConfig": [
			{
				"name": [
					{"service": "user.UserService", "method": "GetUser"}
				],
				"retryPolicy": {
					"maxAttempts": 3,
					"initialBackoff": "0.1s",
					"maxBackoff": "1s",
					"backoffMultiplier": 2.0,
					"retryableStatusCodes": ["UNAVAILABLE","RESOURCE_EXHAUSTED"]
				}
			}
		]
	}`
	conn, err := grpc.Dial(
		"etcd:///"+serviceKey, // 使用 etcd resolver
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("连接 gRPC 服务失败: %v", err)
	}
	log.Printf("grpc conn target:%s\n", conn.Target())
	log.Println("gRPC 连接成功，服务发现完成")

	// 5. 构建 grpc 客户端
	userClient := NewUserServiceClient(conn)

	// 6. 模拟调用 RPC 方法
	log.Println("==================开始发送grpc请求==================")
	for i := 0; i < 10; i++ {
		ctx := context.WithValue(context.Background(), "uid", strconv.Itoa(i))
		_, err := userClient.GetUser(ctx, &GetUserReq{
			Uid: "12345677",
		})
		if err != nil {
			log.Printf("调用 RPC 方法失败: %v\n", err)
		} else {
			log.Printf("调用 RPC 方法成功\n")
		}
		// time.Sleep(1 * time.Second)
	}
	//等待所有请求完成再注销服务
	cancel()
	time.Sleep(5 * time.Second)
	cli.Revoke(context.Background(), lease.ID)
	conn.Close()
}

type server struct {
	UnimplementedUserServiceServer
	index int
}

func (s *server) GetUser(cex context.Context, req *GetUserReq) (*GetUserResp, error) {
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

func startServer(index int, addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	s := grpc.NewServer()
	RegisterUserServiceServer(s, &server{index: index})
	// 创建健康检查服务
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	log.Printf("服务端正在监听: %s\n", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务端启动失败: %v", err)
	}
}
