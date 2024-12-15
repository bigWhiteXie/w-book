package balancer

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
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
			curWeight:  int32(weight),
		}
		nodes = append(nodes, node)
	}
	return &weightedRoundRobinPicker{
		nodes: nodes,
	}
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
		fmt.Printf("{addr:%s, curWeight:%d} ", n.addr, n.curWeight)
	}
	fmt.Printf("\n")
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
				//认为该节点已经崩了
				pickNode.setAvaliable(StatusUnavailable)
			case nil:
				//响应正常
				pickNode.status = StatusAvailable
				pickNode.addCurWeight(1)
			default:
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.Unavailable, codes.ResourceExhausted:
						// 此时大概率是熔断和过载,设置不可用
						log.Printf("服务暂时不可用\n")
						pickNode.setAvaliable(StatusTemporary)
						pickNode.curWeight = minWeightThresoldTimes * pickNode.initWeight
					}
				}
			}
		},
	}, nil
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
