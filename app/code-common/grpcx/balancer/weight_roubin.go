package balancer

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Status int

const (
	StatusUnavailable Status = iota // 0: 不可用
	StatusTemporary                 // 1: 临时可用
	StatusAvailable                 // 2: 可用
)

const (
	weightedRoundRobinName = "weighted_round_robin"
	maxWeightThresoldTimes = 5
	minWeightThresoldTimes = -1
)

func InitWeightRoubin() {
	balancer.Register(base.NewBalancerBuilder(weightedRoundRobinName, &weightedRoundRobinBuilder{}, base.Config{}))
}

// weightedRoundRobinBuilder 实现 balancer.Builder 接口
type weightedRoundRobinBuilder struct{}

func (b *weightedRoundRobinBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	log.Println("服务实例发生变化，重新构造新的picker")
	nodes := make([]*node, 0, len(info.ReadySCs))
	for subconn, info := range info.ReadySCs {
		metadata := info.Address.Metadata.(map[string]interface{})
		weight, err := strconv.Atoi(metadata["weight"].(string))
		if err != nil {
			log.Fatal("pickerBuildInfo中weight字段异常:%s", err)
		}
		node := &node{
			addr:       info.Address.Addr,
			status:     StatusAvailable,
			subconn:    subconn,
			initWeight: int32(weight),

			curWeight: int32(weight),
		}
		nodes = append(nodes, node)
	}
	picker := &weightedRoundRobinPicker{
		nodes: nodes,
	}
	//定期检查熔断服务是否恢复正常
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for _ = range ticker.C {
			picker.lock.RLock()
			for _, node := range picker.nodes {
				if node.status == StatusUnavailable && node.healthyCheck() {
					node.status = StatusTemporary
				}
			}
			picker.lock.RUnlock()
		}

	}()

	return picker
}

// weightedRoundRobinPicker 实现 balancer.Picker 接口
type weightedRoundRobinPicker struct {
	nodes []*node
	lock  sync.RWMutex
}

func (p *weightedRoundRobinPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var (
		pickNode *node
		total    int32
	)
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, node := range p.nodes {
		if node.status == StatusUnavailable {
			continue
		}
		increment := node.initWeight
		if node.status == StatusTemporary {
			increment = increment / 4
		}
		total += node.initWeight
		node.addCurWeight(increment)
		if pickNode == nil || node.curWeight > pickNode.curWeight {
			pickNode = node
		}
	}
	if pickNode == nil {
		return balancer.PickResult{}, fmt.Errorf("no subconns available")
	}
	fmt.Printf("当前节点:")
	for _, n := range p.nodes {
		fmt.Printf("{addr:%s, curWeight:%d}\n", n.addr, n.curWeight)
	}

	log.Printf("负载均衡选中节点:%s,权重:%d\n", pickNode.addr, pickNode.curWeight)
	// 扣减当前权重并确保权重不会小于该节点的最小值
	pickNode.addCurWeight(-total)
	return balancer.PickResult{
		SubConn: pickNode.subconn,
		Done: func(di balancer.DoneInfo) {
			p.lock.Lock()
			defer p.lock.Unlock()
			err := di.Err
			switch err {
			case context.DeadlineExceeded, io.EOF, io.ErrUnexpectedEOF:
				//认为该节点已经崩了,移除node
				p.removeNode(pickNode)
			case nil:
				//响应正常
				pickNode.status = StatusAvailable
				pickNode.addCurWeight(1)
			default:
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.ResourceExhausted:
						log.Printf("节点%v开始限流,降低权重\n", pickNode)
						pickNode.setAvaliable(StatusTemporary)
						pickNode.curWeight = minWeightThresoldTimes * pickNode.initWeight
					case codes.Unavailable:
						log.Printf("节点%v熔断,后续通过健康探测恢复正常\n", pickNode)
						pickNode.setAvaliable(StatusUnavailable)
					}
				}
			}
		},
	}, nil
}

func (p *weightedRoundRobinPicker) removeNode(pickNode *node) {
	index := 0
	for i, v := range p.nodes {
		if v == pickNode {
			index = i
			break
		}
	}
	p.nodes = append(p.nodes[:index], p.nodes[index:]...)
}

type node struct {
	addr       string
	subconn    balancer.SubConn
	initWeight int32
	curWeight  int32
	status     Status
}

func (n *node) setAvaliable(status Status) {
	n.status = status
}

func (n *node) getStatus() Status {
	return n.status
}

func (n *node) addCurWeight(w int32) {
	n.curWeight += w
	if n.curWeight < minWeightThresoldTimes*n.initWeight {
		n.curWeight = minWeightThresoldTimes * n.initWeight
	}
	if n.curWeight > n.initWeight*maxWeightThresoldTimes {
		n.curWeight = n.initWeight * maxWeightThresoldTimes
	}
}

func (n *node) healthyCheck() bool {
	conn, err := grpc.Dial(n.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	healthClient := grpc_health_v1.NewHealthClient(conn)
	if response, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{}); err == nil {
		switch response.Status {
		case grpc_health_v1.HealthCheckResponse_SERVING:
			return true
		case grpc_health_v1.HealthCheckResponse_NOT_SERVING:
			return false
		case grpc_health_v1.HealthCheckResponse_UNKNOWN:
			return false
		}
	}

	return false
}
