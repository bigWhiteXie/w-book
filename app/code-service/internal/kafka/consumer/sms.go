package consumer

import (
	"context"
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"

	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-code/pkg/sms"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

type SmsConsumerGroup struct {
	topic        string
	client       sarama.ConsumerGroup
	interuptChan chan struct{}
	codeRepo     repo.SmsRepo
	once         sync.Once
	pool         *ants.Pool
}

func NewSmsConsumer(topic string, client sarama.ConsumerGroup, codeRepo repo.SmsRepo) *SmsConsumerGroup {
	pool, _ := ants.NewPool(256, ants.WithExpiryDuration(1*time.Second), ants.WithNonblocking(false), ants.WithMaxBlockingTasks(math.MaxInt64))
	return &SmsConsumerGroup{
		topic:        topic,
		client:       client,
		codeRepo:     codeRepo,
		interuptChan: make(chan struct{}, 1),
		pool:         pool,
	}
}

func (s *SmsConsumerGroup) StartConsumer() {
	defer s.client.Close()

	ctx, cancel := context.WithCancel(context.Background())

	for {
		err := s.client.Consume(ctx, []string{s.topic}, s)
		if err != nil {
			logx.Errorf("[sms kafka] fail to kafka msg,cause:%s", err)
		}
		select {
		case <-s.interuptChan:
			cancel()
			logx.Info("[sms-kafka] close gracefully")
			s.pool.Release()
			return
		default:
		}
	}
}

func (s *SmsConsumerGroup) Stop() {
	s.once.Do(func() {
		close(s.interuptChan)
	})
}
func (SmsConsumerGroup) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (SmsConsumerGroup) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h SmsConsumerGroup) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.pool.Submit(func() {
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
				sess.MarkMessage(msg, "")
				return
			}
			if record.Status != 0 { //幂等性校验失败
				logx.Errorf("record[%v] already handled", record)
				sess.MarkMessage(msg, "")
				return
			}

			//调用smsSvc发送短信
			data := make(map[string]string)
			// 将 JSON 反序列化为 map
			json.Unmarshal([]byte(record.Content), &data)
			if err := sms.MustSendSms(ctx, record.Phone, data); err == nil {
				record.Status = 1
			} else {
				record.Status = 4
			}
			h.codeRepo.UpdateById(ctx, record)

			// 标记，sarama会自动进行提交，默认间隔1秒
			sess.MarkMessage(msg, "")
		})
	}
	return nil
}
