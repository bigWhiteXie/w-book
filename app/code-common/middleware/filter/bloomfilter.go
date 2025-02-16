package filter

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spaolacci/murmur3"
)

type FilterRes int

const (
	ValExist    FilterRes = 1
	ValNotExist FilterRes = 0
	KeyNotExist FilterRes = 2
)

type RedisBloomFilter struct {
	client    *redis.Client
	size      uint
	numHashes uint
}

func NewRedisBloomFilter(client *redis.Client, size uint, numHashes uint) *RedisBloomFilter {
	return &RedisBloomFilter{
		client:    client,
		size:      size,
		numHashes: numHashes,
	}
}

// 批量添加元素到布隆过滤器
func (r *RedisBloomFilter) Add(ctx context.Context, key string, expire time.Duration, values ...interface{}) error {
	positions := make([][]uint64, 0, len(values))
	for _, value := range values {
		positions = append(positions, r.getHashPositions(value))
	}

	// 使用pipline批量设置位图
	pipe := r.client.Pipeline()
	for _, pos := range positions {
		for _, p := range pos {
			pipe.SetBit(ctx, key, int64(p), 1)
		}
	}
	if expire > 0 {
		pipe.Expire(ctx, key, expire)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisBloomFilter) AddNoCreate(ctx context.Context, key string, expire time.Duration, values ...interface{}) error {
	// 判断key是否存在
	exists := r.client.Exists(ctx, key)
	if exists.Val() == 0 {
		return nil
	}
	positions := make([][]uint64, 0, len(values))
	for _, value := range values {
		positions = append(positions, r.getHashPositions(value))
	}

	// 使用pipline批量设置位图
	pipe := r.client.Pipeline()
	for _, pos := range positions {
		for _, p := range pos {
			pipe.SetBit(ctx, key, int64(p), 1)
		}
	}
	if expire > 0 {
		pipe.Expire(ctx, key, expire)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisBloomFilter) JudgeKeys(ctx context.Context, keys []string, value interface{}) ([]FilterRes, error) {
	res := make([]FilterRes, len(keys))
	pipe := r.client.Pipeline()

	// 创建命令集合
	existsCmds := make([]*redis.IntCmd, len(keys))

	// 批量发送EXISTS命令
	for i, key := range keys {
		existsCmds[i] = pipe.Exists(ctx, key)
	}

	// 执行EXISTS检查
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	// 确定需要检查的key
	needCheck := make(map[int]struct{})
	for i, cmd := range existsCmds {
		if cmd.Val() == 0 {
			res[i] = KeyNotExist
		} else {
			needCheck[i] = struct{}{}
		}
	}

	// 第二阶段：批量获取位图状态
	positions := r.getHashPositions(value)
	bitCmds := make([][]*redis.IntCmd, len(keys))
	pipe = r.client.Pipeline()
	for i := range keys {
		if _, ok := needCheck[i]; !ok {
			continue
		}
		bitCmds[i] = make([]*redis.IntCmd, len(positions))
		for j, pos := range positions {
			bitCmds[i][j] = pipe.GetBit(ctx, keys[i], int64(pos))
		}
	}

	// 执行并处理结果
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	for i := range keys {
		if _, ok := needCheck[i]; !ok {
			continue
		}
		exist := true
		for _, cmd := range bitCmds[i] {
			if cmd.Val() == 0 {
				exist = false
				break
			}
		}
		if exist {
			res[i] = ValExist
		} else {
			res[i] = ValNotExist
		}
	}

	return res, nil
}

func (r *RedisBloomFilter) getHashPositions(item interface{}) []uint64 {
	// 将任意类型转换为字节序列
	var data []byte
	switch v := item.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		// 使用gob编码处理复杂类型
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(v); err != nil {
			panic(fmt.Sprintf("bloomfilter: unsupported type %T", item))
		}
		data = buf.Bytes()
	}

	// 生成哈希值
	positions := make([]uint64, r.numHashes)
	hash1 := murmur3.Sum64(data)
	hash2 := fnv.New64a()
	hash2.Write(data)
	hash2Val := hash2.Sum64()

	// 计算哈希位置
	for i := uint(0); i < r.numHashes; i++ {
		positions[i] = (hash1 + uint64(i)*hash2Val) % uint64(r.size)
	}
	return positions
}
