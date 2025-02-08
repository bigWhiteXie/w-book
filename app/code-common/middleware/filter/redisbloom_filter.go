package filter

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spaolacci/murmur3"
)

type WindowedRedisBloomFilter struct {
	client     *redis.Client
	keyPrefix  string        // 布隆过滤器key前缀
	windowSize time.Duration // 时间窗口大小
	size       uint
	numHashes  uint
}

// NewWindowedRedisBloomFilter 创建一个基于Redis的滑动窗口布隆过滤器
// client: Redis客户端
// keyPrefix: Redis key前缀
// windowSize: 滑动窗口大小
// size: 布隆过滤器大小
// falsePositiveRate: 可接受的误判率(0-1之间，例如0.01表示1%的误判率)
func NewWindowedRedisBloomFilter(client *redis.Client, keyPrefix string, windowSize time.Duration, size uint, falsePositiveRate float64) *WindowedRedisBloomFilter {
	// 验证误判率范围
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		panic("false positive rate must be between 0 and 1")
	}

	// 计算最优哈希函数数量
	// k = -log2(falsePositiveRate)
	numHashes := uint(math.Ceil(-math.Log2(falsePositiveRate)))

	// 为了防止哈希函数太多影响性能，可以设置一个上限
	const maxHashes = 10
	if numHashes > maxHashes {
		numHashes = maxHashes
	}

	return &WindowedRedisBloomFilter{
		client:     client,
		keyPrefix:  keyPrefix,
		windowSize: windowSize,
		size:       size,
		numHashes:  numHashes,
	}
}

func (w *WindowedRedisBloomFilter) getCurrentKey() string {
	// 使用当前时间窗口作为key的一部分
	window := time.Now().Unix() / int64(w.windowSize.Seconds())
	return fmt.Sprintf("%s:%d", w.keyPrefix, window)
}

func (r *WindowedRedisBloomFilter) getHashPositions(item string) []uint64 {
	positions := make([]uint64, r.numHashes)
	hash1 := murmur3.Sum64([]byte(item))
	hasher := fnv.New64a()
	hasher.Write([]byte(item))
	hash2 := hasher.Sum64()

	for i := uint(0); i < r.numHashes; i++ {
		positions[i] = (hash1 + uint64(i)*hash2) % uint64(r.size)
	}
	return positions
}

func (w *WindowedRedisBloomFilter) Add(ctx context.Context, item string) error {
	currentKey := w.getCurrentKey()
	positions := w.getHashPositions(item)
	pipe := w.client.Pipeline()

	// 设置位图
	for _, pos := range positions {
		pipe.SetBit(ctx, currentKey, int64(pos), 1)
	}

	// 设置过期时间为2个窗口
	pipe.Expire(ctx, currentKey, w.windowSize*2+time.Second)

	_, err := pipe.Exec(ctx)
	return err
}

func (w *WindowedRedisBloomFilter) Test(ctx context.Context, item string) (bool, error) {
	// 检查当前窗口和前一个窗口
	currentWindow := time.Now().Unix() / int64(w.windowSize.Seconds())
	keys := []string{
		fmt.Sprintf("%s:%d", w.keyPrefix, currentWindow),
		fmt.Sprintf("%s:%d", w.keyPrefix, currentWindow-1),
	}

	positions := w.getHashPositions(item)

	// 检查所有窗口
	for _, key := range keys {
		exists, err := w.testKey(ctx, key, positions)

		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}

	return false, nil
}

// 检查指定key的位图是否所有位置的bit都为1
func (w *WindowedRedisBloomFilter) testKey(ctx context.Context, key string, positions []uint64) (bool, error) {
	pipe := w.client.Pipeline()

	// 检查所有位置的bit是否都为1
	for _, pos := range positions {
		pipe.GetBit(ctx, key, int64(pos))
	}

	// 执行管道命令
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("execute redis pipeline failed: %w", err)
	}

	// 如果任何一个位置的bit为0，则元素一定不存在
	for _, cmd := range cmds {
		if cmd.(*redis.IntCmd).Val() == 0 {
			return false, nil
		}
	}

	// 所有位置的bit都为1，元素可能存在
	return true, nil
}
