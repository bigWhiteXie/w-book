package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/panjf2000/ants/v2"

	"codexie.com/w-book-code/internal/logic"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-common/common/codeerr"
	"codexie.com/w-book-common/metric"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

type SmsConsumerGroup struct {
	topic      string
	group      string
	client     sarama.ConsumerGroup
	smsService logic.SmsService
	codeRepo   repo.SmsRepo
	cancel     context.CancelFunc
	pool       *ants.Pool
}

// 用户针对需求对consumerGroup做好配置
// 短信服务依赖该consumerGroup启动消费者
func NewSmsConsumer(topic string, client sarama.ConsumerGroup, codeRepo repo.SmsRepo, smsService logic.SmsService) *SmsConsumerGroup {
	pool, _ := ants.NewPool(256, ants.WithExpiryDuration(1*time.Second), ants.WithNonblocking(false), ants.WithMaxBlockingTasks(math.MaxInt64))
	return &SmsConsumerGroup{
		topic:      topic,
		group:      "sms-service",
		client:     client,
		smsService: smsService,
		codeRepo:   codeRepo,
		pool:       pool,
	}
}

func (s *SmsConsumerGroup) StartConsumer() {
	defer s.client.Close()

	for {
		ctx, cancel := context.WithCancel(context.Background())
		s.cancel = cancel
		err := s.client.Consume(ctx, []string{s.topic}, s)
		if err != nil {
			logx.Errorf("[sms kafka] fail to kafka msg,cause:%s", err)
		}
		select {
		case <-ctx.Done():
			logx.Info("[sms-kafka] close gracefully")
			s.pool.Release()
			return
		default:
		}
	}
}

func (s *SmsConsumerGroup) Stop() {
	s.cancel()
}

func (SmsConsumerGroup) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (SmsConsumerGroup) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h SmsConsumerGroup) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		metric.ConsumeCountCounter.WithLabelValues(h.topic, h.group).Add(1)
		sess.MarkMessage(msg, "")
		h.pool.Submit(func() {
			begin := time.Now()
			logx.Infof("Received message: key=%s, value=%s, partition=%d, offset=%d\n", string(msg.Key), string(msg.Value), msg.Partition, msg.Offset)
			//幂等性校验及状态校验
			record, err := h.codeRepo.FindById(context.Background(), string(msg.Value))
			ctx := context.Background()
			if err != nil {
				logx.Errorf("[sms consumer] fail to consumer msg %s, cause: %v", string(msg.Value), err)
				if record != nil {
					record.ErrorMsg = err.Error()
					h.codeRepo.UpdateById(ctx, record)
				}
				return
			}
			if record.Status != 0 { //幂等性校验失败
				logx.Errorf("[sms consumer] 幂等性校验失败，该记录已经被处理过,record=%v", record)
				return
			}

			//调用smsSvc发送短信
			data := make(map[string]string)
			json.Unmarshal([]byte(record.Content), &data)
			for {
				var err error
				codeErr := codeerr.WithCodeErr{}
				if _, err = h.smsService.SendSms(ctx, record.Phone, data); err == nil {
					record.Status = 1
					h.codeRepo.UpdateById(ctx, record)
					metric.ConsumeTimeHistogram.WithLabelValues(h.topic, h.group).Observe(float64(time.Since(begin).Milliseconds()))
					return
				}
				if errors.As(err, codeErr) && codeErr.Code == codeerr.SmsNotAvaliableErr {
					metric.ConsumeErrCounter.WithLabelValues(h.topic, h.group, err.Error()).Inc()
					logx.Errorf("[SmsConsumerGroup] 发送短信[id=%d]时无短信服务商正常运行")
					return
				}
				logx.Errorf("[SmsConsumerGroup] 异步发送短信异常，尝试再次发送:%s", err)
			}
		})
	}
	return nil
}
