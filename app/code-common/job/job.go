package job

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

type jobRun func()

type Job interface {
	Name() string
	Run() error
	TimeExper() string
}

type JobBuilder struct {
	cron *cron.Cron
}

func NewJobBuilder(c *cron.Cron) *JobBuilder {
	bd := &JobBuilder{cron: c}
	return bd
}

func (b *JobBuilder) AddJob(job Job, executeNow bool) error {
	run := b.build(job, executeNow)
	_, err := b.cron.AddFunc(job.TimeExper(), run)

	return err
}
func (b *JobBuilder) build(job Job, executeNow bool) jobRun {
	if executeNow {
		job.Run()
	}
	return func() {
		start := time.Now()
		logx.Infof("任务开始 %s  %s", job.Name(), start.String())
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
	logx.Infof("job完成退出,耗时%f秒", time.Since(now).Seconds())
}
