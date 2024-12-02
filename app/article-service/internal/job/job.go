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

func InitJobBuilder(rankJob *RankingJob) *JobBuilder {
	c := cron.New()
	bd := &JobBuilder{cron: c}

	run := bd.build(rankJob)
	c.AddFunc(rankJob.TimeExper(), run)

	return bd
}

func (b *JobBuilder) build(job Job) jobRun {
	job.Run()
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
