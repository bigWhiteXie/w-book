package rpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	etcdAddress = "localhost:2479" // etcd服务地址
	serviceName = "test.rpc"
	serviceKey  = "test/rpc" // 服务注册前缀
	nodeAddress = "localhost:50051"
	nodeName    = "test-node"
	dialTimeout = 5 * time.Second
	leaseTTL    = 10 // 租约时间（秒）
)

// 测试etcd的服务注册、监听与发现
// em就是用来管理服务实例的注册、注销、监听
func TestEtcdRegistry(t *testing.T) {
	//==========================模拟服务注册==========================
	// 1. 连接 etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2479"}, // etcd 地址
		DialTimeout: 5 * time.Second,            // 连接超时时间
	})
	if err != nil {
		t.Fatalf("连接 etcd 失败: %v", err)
	}
	defer cli.Close()

	// 2. 创建 endpoint.Manager(针对某服务简化管理)，监听某服务下所有实例的变更
	em, err := endpoints.NewManager(cli, serviceKey)
	if err != nil {
		t.Fatalf("创建 endpoint.Manager 失败: %v", err)
	}

	// 3. 服务发布
	serviceEndpoint := "127.0.0.1:8080"
	lease, err := cli.Grant(context.Background(), 10) // 10 秒的租约

	if err != nil {
		t.Fatalf("创建租约失败: %v", err)
	}
	// 将服务端点注册到 etcd
	endpoint := endpoints.Endpoint{
		Addr: serviceEndpoint,
		Metadata: map[string]string{
			"weight": "10",
			"cpu":    "80%",
			"memory": "20%",
		},
	}

	//将服务端点注册在serviceKey下，并绑定租约
	err = em.AddEndpoint(context.Background(), serviceKey+"/"+strconv.Itoa(int(lease.ID)), endpoint, clientv3.WithLease(lease.ID))
	if err != nil {
		t.Fatalf("服务发布失败: %v", err)
	}
	log.Printf("服务发布成功: %s -> %s\n", serviceKey, serviceEndpoint)
	//==========================服务注册结束==========================

	kvMap, _ := em.List(context.Background())
	log.Println("============em拉取服务实例列表============")
	for k, v := range kvMap {
		log.Printf("%s -> %s\n", k, v)
	}
	log.Println("============em拉取服务实例列表结束============")
	//==========================服务监听==========================
	go func() {
		ch, _ := em.NewWatchChannel(context.Background())
		for {
			select {
			case events, ok := <-ch:
				if !ok {
					log.Println("endpoint manager 的channel关闭，退出监听")
					return
				}
				for _, evt := range events {
					log.Printf("监听事件：服务%s，Op:%v, key:%s, addr:%s, metadata:%v\n", serviceKey, evt.Op, evt.Key, evt.Endpoint.Addr, evt.Endpoint.Metadata)
				}
			}
		}
	}()
	time.Sleep(2 * time.Second)
	for i := 0; i < 3; i++ {
		err = em.AddEndpoint(context.Background(), serviceKey+"/"+strconv.Itoa(i), endpoint, clientv3.WithLease(lease.ID))
		if err != nil {
			log.Printf("新增服务实例失败,原因:%s\n", err)
		}
	}
	// 5. 模拟服务续约
	go func() {
		for {
			time.Sleep(5 * time.Second) // 每 5 秒续约一次
			_, err := cli.KeepAliveOnce(context.Background(), lease.ID)
			if err != nil {
				log.Printf("续约失败: %v", err)
				return
			}
			log.Println("续约成功")
		}
	}()

	// 6. 模拟服务运行
	time.Sleep(60 * time.Second) // 模拟服务运行 30 秒

	// 7. 服务注销
	err = em.DeleteEndpoint(context.Background(), serviceKey+"/")
	if err != nil {
		t.Fatalf("服务注销失败: %v", err)
	}
	log.Println("服务注销成功")
	kvMap, _ = em.List(context.Background())
	log.Println("============em拉取服务实例列表============")
	for k, v := range kvMap {
		log.Printf("%s -> %s\n", k, v)
	}
	log.Println("测试结束")
}

func TestEtcdDel(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2479"}, // etcd 地址
		DialTimeout: 5 * time.Second,            // 连接超时时间
	})
	if err != nil {
		t.Fatalf("连接 etcd 失败: %v", err)
	}
	defer cli.Close()

	em, err := endpoints.NewManager(cli, serviceKey)
	for i := 0; i < 3; i++ {
		em.DeleteEndpoint(context.TODO(), serviceKey+"/"+strconv.Itoa(i))
	}
}

// 服务发现
func TestGrpcDiscover(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddress},
		DialTimeout: dialTimeout,
	})
	if err != nil {
		panic(err)
	}
	//解析url获取对应的实例地址
	etcdResolverBuilder, _ := resolver.NewBuilder(cli)
	conn, err := grpc.Dial("etcd:///code.rpc", grpc.WithResolvers(etcdResolverBuilder), grpc.WithTransportCredentials(insecure.NewCredentials()))
	log.Printf("grpc clientConn state =>:%v", conn.GetState())
}

// RegisterService 直接使用etcd client注册服务到etcd
func RegisterService(cli *clientv3.Client, key, value string) error {
	// 创建租约
	resp, err := cli.Grant(context.Background(), leaseTTL)
	if err != nil {
		return err
	}

	// 注册服务
	_, err = cli.Put(context.Background(), key, value, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	// 保持租约
	_, err = cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Service registered: %s -> %s\n", key, value)
	return nil
}

// QueryService 查询指定前缀下的所有服务
func QueryService(cli *clientv3.Client, prefix string) ([]string, error) {
	resp, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var services []string
	for _, kv := range resp.Kvs {
		services = append(services, fmt.Sprintf("%s -> %s", kv.Key, kv.Value))
	}
	return services, nil
}

// StartGRPCServer 启动一个简单的gRPC服务
func StartGRPCServer() {
	listener, err := net.Listen("tcp", nodeAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	log.Printf("gRPC server started at %s", nodeAddress)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
