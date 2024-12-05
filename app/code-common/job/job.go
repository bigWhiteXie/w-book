package job

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
	"k8s.io/apimachinery/pkg/util/rand"
)

var (
	jobLockPrefix  = "job:lock:"
	jobLoadPrefix  = "job:load:"
	jobAlivePrefix = "job:alive:"
	jobSpecPrefix  = "job:spec:"
)

var (
	maxLoadThreshold int32 = 85
	minLoadThreshold int32 = 60
	overLoadTimes    int32 = 0
)

var (
	OverLoadErr = errors.New("loadbalance")
)

type jobRun func()

type Job interface {
	Name() string
	Run() error
	TimeExper() string
}

type JobBuilder struct {
	Name string
	Id   string

	loadScore    int32
	cron         *cron.Cron
	lockClient   *rlock.Client
	redisClient  *redis.Client
	timeout      time.Duration
	ticker       *time.Ticker
	localLockMap map[string]sync.Mutex
	redLockMap   map[string]*rlock.Lock
}

func NewJobBuilder(c *cron.Cron, redisClient *redis.Client, name string, timeout time.Duration) *JobBuilder {
	Id := uuid.New().String()[:16]
	rs := rlock.NewClient(redisClient)
	ticker := time.NewTicker(30 * time.Second)
	bd := &JobBuilder{cron: c, Id: Id, ticker: ticker, Name: name, lockClient: rs, redisClient: redisClient, localLockMap: make(map[string]sync.Mutex), redLockMap: make(map[string]*rlock.Lock), timeout: timeout}
	go func() {
		for _ = range ticker.C {
			score := int32(bd.computeLoadBalance())
			atomic.StoreInt32(&bd.loadScore, score)
			logx.Infof("[%s] 当前负载分数:%d", name+"-"+Id, bd.loadScore)
			// 向zset发送负载均衡并保持存活状态
			if err := redisClient.ZAdd(context.Background(), jobLoadPrefix+name, redis.Z{Score: float64(bd.loadScore), Member: Id}).Err(); err != nil {
				logx.Errorf("Job[%s]更新负载分数失败:%s", err)
			}
			if err := redisClient.Set(context.Background(), jobAlivePrefix+Id, 1, 40*time.Second).Err(); err != nil {
				logx.Errorf("Job[%s]保持心跳失败:%s", err)
			}
			if bd.loadScore > maxLoadThreshold {
				atomic.AddInt32(&overLoadTimes, 1)
			} else {
				atomic.StoreInt32(&overLoadTimes, 0)
			}
		}
	}()
	return bd
}

func (b *JobBuilder) AddJob(job Job, executeNow bool) error {
	run := b.build(job, executeNow)
	_, err := b.cron.AddFunc(job.TimeExper(), run)

	return err
}
func (b *JobBuilder) build(job Job, executeNow bool) jobRun {
	var (
		localLock = sync.Mutex{}
	)
	if executeNow {
		job.Run()
	}
	b.localLockMap[job.Name()] = localLock
	return func() {
		//续约和定时任务可能同时访问共享资源，因此使用锁
		localLock.Lock()
		lock, ok := b.redLockMap[job.Name()]
		localLock.Unlock()

		//分布式锁不存在，尝试获取锁
		if !ok || lock == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			lock, err := b.tryLockOnOverload(ctx, job.Name())
			if err != nil {
				logx.Errorf("[%s] 获取分布式锁失败, 丢弃该任务", job.Name(), err)
				return
			}
			//开启协程进行续约，考虑协程主动退出
			go func() {
				//阻塞操作，锁释放或则redis连接不上会退出
				//当redis连接不上时会取消续期，就算此时锁没过期，等过期后还是能够被争抢
				err := lock.AutoRefresh(b.timeout/2, 2*time.Second)
				if err != nil {
					logx.Errorf("[%s] 续约锁失败:%s", job.Name(), err)
				}
				b.releaseLock(job.Name())
			}()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		defer b.unLockOnOverload(ctx, job.Name(), job.TimeExper())
		start := time.Now()
		logx.Infof("任务开始 %s  %s", job.Name(), start.String())
		// 分布式锁抢占任务，抢占失败直接return
		err := job.Run()
		if err != nil {
			logx.Errorf("任务[%s]执行失败:%s", job.Name(), err)
		} else {
			logx.Infof("任务[%s]执行结束，耗时 %f 秒", job.Name(), time.Since(start).Seconds())
		}
	}
}

func (b *JobBuilder) tryLockOnOverload(ctx context.Context, jobName string) (*rlock.Lock, error) {
	if overLoadTimes >= 3 {
		return nil, OverLoadErr
	}
	localLock := b.localLockMap[jobName]
	localLock.Lock()
	defer localLock.Unlock()
	//todo 若该任务存在指定实例且不是当前实例则放弃
	id, err := b.redisClient.Get(ctx, jobSpecPrefix+jobName).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if id != "" && id != b.Id {
		return nil, errors.New(fmt.Sprintf("job[%s]已经指定实例%s", jobName, b.Name+id))
	}

	lock, err := b.lockClient.Lock(ctx, jobLockPrefix+jobName, b.timeout, &rlock.FixIntervalRetry{
		Interval: time.Second,
		Max:      3,
	}, time.Second)
	if err != nil {
		return nil, err
	}
	logx.Infof("[%s] 抢占到job %s 的分布式锁", b.Name+b.Id, jobName)
	b.redLockMap[jobName] = lock
	return lock, nil
}

func (b *JobBuilder) unLockOnOverload(ctx context.Context, jobName string, express string) error {
	var (
		zsetKey   = jobLoadPrefix + b.Name
		localLock = b.localLockMap[jobName]
	)

	localLock.Lock()
	redLock, ok := b.redLockMap[jobName]
	localLock.Unlock()
	if !ok {
		logx.Infof("[%s] job %s 已经释放了分布式锁", b.Name+b.Id, jobName)
		return nil
	}

	if atomic.LoadInt32(&overLoadTimes) >= 3 {
		logx.Infof("[%s] 当前实例%s负载过大,释放job", jobName, b.Name+"-"+b.Id)
		return redLock.Unlock(ctx)
	}
	// 从zset中取出所有元素及分数，由低到高排列，进行遍历
	zRangeByScore := &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  10, // 限制数量，防止大范围查询导致性能问题
	}

	elements, err := b.redisClient.ZRangeByScoreWithScores(ctx, zsetKey, zRangeByScore).Result()
	if err != nil {
		redLock.Unlock(ctx)
		return errors.Wrap(err, "failed to fetch zset elements from Redis")
	}
	logx.Debugf("[%s] 当前job多个实例的负载情况:%v", jobName, elements)
	// 若元素等于自身id则直接返回nil，否则比较其是否小于自身负载分数超过20分，是的话则释放锁
	for _, elem := range elements {
		instanceID := elem.Member.(string)
		instanceScore := elem.Score

		if instanceID == b.Id {
			return nil
		}
		logx.Debugf("[%s]判断实例%s是否存活,key:%s", b.Name+b.Id, instanceID, jobAlivePrefix+instanceID)
		score := atomic.LoadInt32(&b.loadScore)
		if err := b.redisClient.Get(ctx, jobAlivePrefix+instanceID).Err(); score-int32(instanceScore) > 0 && err == nil {
			logx.Debugf("[%s] 发现实例%s负载很低, 准备释放锁让其抢占", jobName, instanceID)
			if err := b.releaseLock(jobName); err != nil {
				return err
			}
			schedule, _ := cron.ParseStandard(express)
			now := time.Now()
			nextTime := schedule.Next(now)
			if err := b.redisClient.Set(ctx, jobSpecPrefix+jobName, instanceID, nextTime.Sub(now)); err != nil {
				logx.Errorf("[%s] 指定实例%s执行任务%s失败", b.Name, b.Name+instanceID, jobName)
			}
			return nil
		} else if err == redis.Nil {
			logx.Debugf("[%s] 实例%s不存活了", b.Name+b.Id, b.Name+instanceID)
			b.redisClient.ZRem(ctx, zsetKey, instanceID) //清空不存活的实例
		} else if err != nil {
			redLock.Unlock(ctx)
			logx.Errorf("[%s] 从redis中获取实例 %s 存活状态失败:%s", jobName, b.Name+b.Id, err)
			return err
		}
	}

	return nil
}

func (b *JobBuilder) computeLoadBalance() int {
	return rand.Intn(30)
}

func (b *JobBuilder) releaseLock(jobName string) error {
	localLock := b.localLockMap[jobName]
	localLock.Lock()
	defer localLock.Unlock()
	redLock, ok := b.redLockMap[jobName]
	if !ok {
		logx.Debugf("[%s] job(%s)分布式锁已经被释放", b.Name, jobName)
		return nil
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	if err := redLock.Unlock(ctx); err != nil {
		logx.Debugf("[%s] job(%s)释放分布式锁失败:%s", b.Name, jobName, err)
		return err
	}
	delete(b.redLockMap, jobName)
	return nil
}

func (b *JobBuilder) Start() {
	b.cron.Start()
}

func (b *JobBuilder) Stop() {
	now := time.Now()
	logx.Info("==========job准备退出==============")
	ctx := b.cron.Stop()
	<-ctx.Done()
	b.ticker.Stop()
	for name, lock := range b.redLockMap {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelFunc()
		if lock != nil {
			lock.Unlock(ctx)
		}
		logx.Info("关闭job[%s]", name)
	}

	logx.Infof("job完成退出,耗时%f秒", time.Since(now).Seconds())
}
