package limiter

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// 初始化：读取配置中得到桶容量capacity、每秒发放的数量为tokens
// 限流：限流key limit:rate:{path}   时间戳key limit:time:{path}   实例个数key limit:instance:nummber
//
//		1.判断本地令牌使用原子操作-1，若大于等于0则放行，小于0则进入如下步骤
//		2.本地不够则执行lua脚本获取批量令牌take_tokens=(tokens/(实例个数*2)) ，逻辑如下
//			2.1 根据当前时间戳(秒)和时间戳key(秒级)的差值计算当前应该加上多少令牌put_tokens，若时间戳key不存在则初始化(限流key的值设置为tokens、时间戳key设置为当前时间戳),
//			2.2 根据限流key拿到剩余的令牌rest_tokens, 计算当前令牌个数为current_tokens = max(capacity, rest_tokens + put_tokens)
//			2.3 若current_tokens <= take_tokens  设置限流key的值为0，并返回current_tokens
//			2.4 若current_tokens >= take_tokens  设置限流key的值为current_tokens - take_tokens，并返回take_tokens
//	    3.拿到take_tokens后使用原子操作重置本地令牌数量，并进行原子操作-1，大于等于0则放行，否则拒绝
type TokenBucketRateConf struct {
	Capacity        int64  `json:",optional"`
	TokensPerSecond int64  `json:",optional"`
	Biz             string `json:",optional"`
}

type TokenBucketLimiter struct {
	redisClient *redis.Client
	mutex       sync.Mutex

	localTokens     int64
	capacity        int64
	updateTime      int64
	tokensPerSecond int64
	instanceCount   int64
	biz             string

	instanceKey      string
	rateKey          string
	timeKey          string
	instanceCountKey string
}

// NewTokenBucketLimiter 初始化限流器

// NewTokenBucketLimiter 函数更新
func NewTokenBucketLimiter(redisClient *redis.Client, conf *TokenBucketRateConf) *TokenBucketLimiter {
	logx.Infof("初始化限流器， tokensPerSecond:%d, capacity:%d, ")
	limiter := &TokenBucketLimiter{
		redisClient: redisClient,
		localTokens: 0, // 本地初始令牌数量

		capacity:        conf.Capacity,
		tokensPerSecond: conf.TokensPerSecond,
		instanceCount:   1,

		rateKey:          fmt.Sprintf("limit:rate:%s", conf.Biz),
		timeKey:          fmt.Sprintf("limit:time:%s", conf.Biz),
		instanceCountKey: fmt.Sprintf("limit:instance:count:%s", conf.Biz),
	}

	// 获取并更新实例数量
	err := limiter.UpdateInstanceCount(context.Background(), true)
	if err != nil {
		logx.Errorf("获取实例数量失败: %v", err)
	}

	return limiter
}

func (l *TokenBucketLimiter) UpdateInstanceCount(ctx context.Context, increment bool) error {
	var err error
	if increment {
		err = l.redisClient.Incr(ctx, l.instanceCountKey).Err()
	} else {
		err = l.redisClient.Decr(ctx, l.instanceCountKey).Err()
	}
	if err != nil {
		logx.Errorf("[TokenBucketLimiter_asdffvaf]更新实例数量失败: %v", err)
		return err
	}
	return nil
}

// Allow 是否允许通过
func (l *TokenBucketLimiter) Allow(ctx context.Context) bool {
	// Step 1: 本地令牌计数原子操作
	if atomic.AddInt64(&l.localTokens, -1) >= 0 {
		logx.Infof("[Allow] 本地令牌限流通过,剩余令牌:%d", l.localTokens)
		// 本地令牌足够，放行
		return true
	}

	// Step 2: 获取锁避免多协程同时向 Redis 请求
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Step 3: 再次检查本地令牌池，避免重复获取
	now := time.Now().Unix()
	if l.updateTime == now {
		if l.localTokens = l.localTokens - 1; l.localTokens >= 0 {
			return true
		}
		return false
	}
	l.updateTime = now
	// Step 4: 本地令牌不足，通过 Redis 获取批量令牌
	takeTokens := l.tokensPerSecond
	return l.fetchTokensFromRedis(ctx, takeTokens)
}

// fetchTokensFromRedis 从 Redis 中获取令牌
func (l *TokenBucketLimiter) fetchTokensFromRedis(ctx context.Context, takeTokens int64) bool {

	luaScript := `
		-- 定义传入的参数
		local rateKey = KEYS[1]
		local timeKey = KEYS[2]
		local instanceCountKey = KEYS[3]
		local currentTime = tonumber(ARGV[1])
		local capacity = tonumber(ARGV[2])
		local tokensPerSecond = tonumber(ARGV[3])
		local takeTokens = tonumber(ARGV[4])
		
		-- 获取 Redis 中的剩余令牌和上次更新时间
		local lastTokens = tonumber(redis.call('get', rateKey)) or capacity
		local lastTime = tonumber(redis.call('get', timeKey)) or currentTime

		-- 获取当前实例的个数
		local instanceCount = tonumber(redis.call('get', instanceCountKey)) or 1
		-- 计算自上次更新以来应增加的令牌数
		local deltaTime = currentTime - lastTime
		local putTokens = deltaTime * tokensPerSecond
		local currentTokens = math.min(capacity, lastTokens + putTokens)

		-- 计算每个实例应该获取的令牌数
		local tokensPerInstance = takeTokens / instanceCount
		-- 更新 Redis 中的令牌数量和时间戳
		if currentTokens < tokensPerInstance then
			redis.call('set', rateKey, 0)
			redis.call('set', timeKey, currentTime)
			return currentTokens
		else
			redis.call('set', rateKey, currentTokens - tokensPerInstance)
			redis.call('set', timeKey, currentTime)
			return tokensPerInstance
		end
	`

	// 执行 Lua 脚本
	result, err := l.redisClient.Eval(ctx, luaScript, []string{l.rateKey, l.timeKey, l.instanceCountKey}, l.updateTime, l.capacity, l.tokensPerSecond, takeTokens).Result()

	if err != nil {
		logx.Errorf("[TokenBucketLimiter_asdasd] 限流器从redis获取令牌失败,cause:", err)
		return false
	}

	tokens, ok := result.(int64)
	l.localTokens = tokens
	logx.Infof("[TokenBucketLimiter_DSGdgew] 拿到令牌数量：%d", tokens)
	if !ok || tokens <= 0 {
		// Redis 中没有足够的令牌
		logx.Alert("[TokenBucketLimiter_fdzggwg] 触发限流")
		return false
	}

	// Step 3: 成功获取 Redis 令牌，更新本地令牌数量
	l.localTokens = tokens - 1
	logx.Infof("[Allow] 限流通过,剩余令牌:%d", l.localTokens)

	return true
}
