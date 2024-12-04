package job

import (
	"context"
	"sync"
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	jobLockPrefix = "job:lock:"
)

type jobRun func()

type Job interface {
	Name() string
	Run() error
	TimeExper() string
}

type JobBuilder struct {
	cron         *cron.Cron
	lockClient   *rlock.Client
	Name         string
	timeout      time.Duration
	localLockMap map[string]sync.Mutex
	redLockMap   map[string]*rlock.Lock
}

func NewJobBuilder(c *cron.Cron, rs *rlock.Client, name string, timeout time.Duration) *JobBuilder {
	bd := &JobBuilder{cron: c, Name: name, lockClient: rs, localLockMap: make(map[string]sync.Mutex), redLockMap: make(map[string]*rlock.Lock), timeout: timeout}
	return bd

}

func (b *JobBuilder) AddJob(job Job, executeNow bool) error {
	run := b.build(job, executeNow)
	_, err := b.cron.AddFunc(job.TimeExper(), run)

	return err
}
func (b *JobBuilder) build(job Job, executeNow bool) jobRun {
	var (
		lockKey   = jobLockPrefix + job.Name()
		localLock = sync.Mutex{}
	)
	if executeNow {
		job.Run()
	}
	b.localLockMap[job.Name()] = localLock
	return func() {
		//续约协程可能和定时任务同时访问redLockMap
		localLock.Lock()
		lock := b.redLockMap[job.Name()]
		localLock.Unlock()
		//分布式锁不存在，尝试获取锁
		if lock == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			lock, err := b.lockClient.Lock(ctx, lockKey, b.timeout, &rlock.FixIntervalRetry{
				Interval: time.Second,
				Max:      3,
			}, time.Second)
			if err != nil {
				logx.Errorf("[%s] 获取分布式锁失败, 丢弃该任务", job.Name(), err)
				return
			}
			localLock.Lock()
			b.redLockMap[job.Name()] = lock
			localLock.Unlock()
			//开启协程进行续约，考虑协程主动退出
			go func() {
				err := lock.AutoRefresh(b.timeout/2, 2*time.Second)
				if err != nil {
					logx.Errorf("[%s] 续约锁失败:%s", job.Name(), err)
				}
				localLock.Lock()
				b.redLockMap[job.Name()] = nil
				localLock.Unlock()
			}()
		}
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

func (b *JobBuilder) Start() {
	b.cron.Start()
}

func (b *JobBuilder) Stop() {
	now := time.Now()
	logx.Info("==========job准备退出==============")
	ctx := b.cron.Stop()
	<-ctx.Done()

	for name, lock := range b.redLockMap {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelFunc()
		lock.Unlock(ctx)
		logx.Info("关闭job[%s]", name)
	}
	logx.Infof("job完成退出,耗时%f秒", time.Since(now).Seconds())
}
