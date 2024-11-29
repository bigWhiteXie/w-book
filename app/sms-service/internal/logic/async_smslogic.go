package logic

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"codexie.com/w-book-code/internal/model"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-common/producer"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	topic    = "sms-topic"
	lateSize = 100
)

type ASyncSmsLogic struct {
	smsService    SmsService
	sendLatencies []int64
	kafkaProducer producer.Producer
	smsRepo       repo.SmsRepo
	lock          sync.RWMutex
}

func NewASyncSmsLogic(smsService SmsService, smsRepo repo.SmsRepo) *ASyncSmsLogic {
	return &ASyncSmsLogic{
		smsService:    smsService,
		smsRepo:       smsRepo,
		sendLatencies: make([]int64, 0, lateSize),
	}
}

// SendSms实现短信发送逻辑，根据延迟中位数决定发送方式
func (s *ASyncSmsLogic) SendSms(ctx context.Context, phone string, args map[string]string) error {
	//短信记录落库
	record := model.NewSmsRecord(phone, args)
	if err := s.smsRepo.SaveSmsRecord(ctx, record); err != nil {
		return err
	}

	s.lock.RLock()
	medianLatency := s.calculateMedianLatency()
	if medianLatency > 200 {
		logx.Infof("[ASyncSmsLogic] 当前短信发送延迟大于200ms, 改为异步发送短信[id=%d]", record.ID)
		return s.kafkaProducer.SendSync(ctx, topic, strconv.Itoa(record.ID))
	}
	s.lock.RUnlock()

	start := time.Now()
	if err := s.smsService.SendSms(ctx, phone, args); err != nil {
		logx.Errorf("[ASyncSmsLogic] 同步发送短信[id=%d]异常改为异步投送：%s", record.ID, err)
		return s.kafkaProducer.SendSync(ctx, topic, strconv.Itoa(record.ID))
	}

	latency := time.Since(start).Milliseconds()
	s.lock.Lock()
	s.sendLatencies = append(s.sendLatencies, latency)
	if len(s.sendLatencies) > 100 {
		s.sendLatencies = s.sendLatencies[1:]
	}
	s.lock.Unlock()

	return nil
}

// calculateMedianLatency计算短信发送延迟的中位数
func (s *ASyncSmsLogic) calculateMedianLatency() int64 {
	latenciesCopy := make([]int64, 0, len(s.sendLatencies))
	copy(latenciesCopy, s.sendLatencies)
	sort.Slice(latenciesCopy, func(i, j int) bool {
		return latenciesCopy[i] < latenciesCopy[j]
	})

	n := len(latenciesCopy)
	if n%2 == 0 {
		return (latenciesCopy[n/2-1] + latenciesCopy[n/2]) / 2
	} else {
		return latenciesCopy[n/2]
	}
}
