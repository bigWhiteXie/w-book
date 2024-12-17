package balancer

import (
	"fmt"
	"hash/crc32"
	"log"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const (
	consistanceName = "consistance_hash_balance"
)

func InitConsistentBalancer(key string) {
	balancer.Register(base.NewBalancerBuilder(consistanceName, &consistenceMapBuilder{key: key}, base.Config{}))
}

// weightedRoundRobinBuilder 实现 balancer.Builder 接口
type consistenceMapBuilder struct {
	key string
}

func (b *consistenceMapBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	var (
		subconnMap = make(map[string]balancer.SubConn, len(info.ReadySCs))
		addrs      = make([]string, 0, len(info.ReadySCs))
	)
	log.Println("服务实例发生变化，重新构造新的picker")
	for subconn, info := range info.ReadySCs {
		addr := info.Address.Addr
		addrs = append(addrs, addr)
		subconnMap[addr] = subconn
	}
	return &consistenceMapPicker{
		key:            b.key,
		subconns:       subconnMap,
		consistenceMap: NewConsistenceMap(addrs, nil),
	}
}

type consistenceMapPicker struct {
	consistenceMap *ConsistenceMap
	subconns       map[string]balancer.SubConn
	lock           sync.RWMutex
	key            string
}

func (p *consistenceMapPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	val := info.Ctx.Value(p.key).(string)
	addr := p.consistenceMap.Get(val)
	subconn := p.subconns[addr]
	log.Printf("当前uid:%s,hash(uid): %d\n", val, crc32.ChecksumIEEE([]byte(val)))
	fmt.Printf("当前节点:")
	for addr, _ := range p.subconns {
		fmt.Printf("{addr:%s, hashVal:%d} ", addr, crc32.ChecksumIEEE([]byte(addr)))
	}
	fmt.Printf("\n")

	log.Printf("负载均衡选中节点:%s\n", addr)
	// 扣减当前权重并确保权重不会小于该节点的最小值
	return balancer.PickResult{
		SubConn: subconn,
	}, nil
}

type Hash func(data []byte) uint32

// 支持动态增加/减少节点、
type ConsistenceMap struct {
	hash     Hash
	hashKeys []uint32
	keyMap   map[uint32]string
}

// 构建时传入所有的key，建造hash环
func NewConsistenceMap(keys []string, fn Hash) *ConsistenceMap {
	var (
		keyMap   = make(map[uint32]string, len(keys))
		hashKeys = make([]uint32, 0, len(keys))
	)
	if fn == nil {
		fn = crc32.ChecksumIEEE
	}

	for _, key := range keys {
		hashVal := fn([]byte(key))
		keyMap[hashVal] = key
		hashKeys = insertSorted(hashKeys, hashVal)
	}

	return &ConsistenceMap{
		hash:     fn,
		keyMap:   keyMap,
		hashKeys: hashKeys,
	}
}

// 得到对应的节点
func (m *ConsistenceMap) Get(key string) string {
	hashVal := m.hash([]byte(key))
	low, high := 0, len(m.hashKeys)-1
	//找到第一个大于等于hashVal的key
	for low <= high {
		mid := low + (high-low)/2
		midVal := m.hashKeys[mid]
		if midVal == hashVal {
			return m.keyMap[midVal]
		} else if midVal > hashVal {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	if low >= len(m.hashKeys) {
		low = 0
	}

	return m.keyMap[m.hashKeys[low]]
}

func findInsertIndex(arr []uint32, value uint32) int {
	low, high := 0, len(arr)-1
	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] == value {
			return mid // 如果值已存在，返回当前位置
		} else if arr[mid] < value {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return low // 返回插入位置
}

// 插入元素并保持有序
func insertSorted(arr []uint32, value uint32) []uint32 {
	index := findInsertIndex(arr, value)
	arr = append(arr, 0)             // 扩展切片
	copy(arr[index+1:], arr[index:]) // 后移元素
	arr[index] = value               // 插入新元素
	return arr
}
