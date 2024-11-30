package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"codexie.com/w-book-common/metric"
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
)

var (
	createConsumer = "create-evt-consumer-group"
)

// ：kafka-topics.sh --bootstrap-server 192.168.126.100:9092 --topic create-evt-topic --create --partitions 1 --replication-factor 1
type CreateEventListener struct {
	client       sarama.ConsumerGroup
	interactRepo repo.IInteractRepo
	cancel       context.CancelFunc
	once         sync.Once
}

func NewCreateEventListener(config config.Config, interactRepo repo.IInteractRepo) *CreateEventListener {
	conf := config.KafkaConf
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V2_1_0_0
	client, err := sarama.NewConsumerGroup(conf.Brokers, createConsumer, saramaConf)
	if err != nil {
		panic(fmt.Sprintf("unable to create kafka consumer group, cause:%s", err))
	}
	return &CreateEventListener{
		client:       client,
		interactRepo: interactRepo,
	}
}

func (s *CreateEventListener) StartListner() {

	go func() {
		defer s.client.Close()
		ctx, cancel := context.WithCancel(context.Background())
		s.cancel = cancel
		err := s.client.Consume(ctx, []string{domain.CreateEvtTopic}, s)
		if err != nil {
			logx.Errorf("[CreateEventListener] fail to kafka msg,cause:%s", err)
		}
		logx.Infof("[CreateEventListener] start successfully, topic:%s", domain.CreateEvtTopic)
		select {
		case <-ctx.Done():
			logx.Info("[CreateEventListener] close gracefully")
			return
		}
	}()
}

func (s *CreateEventListener) Stop() {
	s.once.Do(func() {
		s.cancel()
	})
}
func (CreateEventListener) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (CreateEventListener) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h CreateEventListener) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	done := false
	for !done {
		errGroup, _ := errgroup.WithContext(context.Background())
		for i := 0; i < batchSize; i++ {
			msg, ok := <-claim.Messages()
			startTime := time.Now()
			metric.ConsumeCountCounter.WithLabelValues(domain.CreateEvtTopic, "interact-group").Inc()
			if !ok {
				done = true
				break
			}
			sess.MarkMessage(msg, "")
			errGroup.Go(func() error {
				defer func() {
					metric.ConsumeTimeHistogram.WithLabelValues(domain.CreateEvtTopic, "interact-group").Observe(float64(time.Since(startTime).Milliseconds()))
				}()
				logx.Infof("添加msg:%s", msg.Value)
				readEvt, _ := convertMsg2ReadEvt(msg)
				if err := h.interactRepo.CreateInteractData(context.Background(), &readEvt); err != nil {
					logx.Errorf("[CreateListener] 消费消息异常:%s", err)
					metric.ConsumeErrCounter.WithLabelValues(domain.CreateEvtTopic, "interact-group", err.Error()).Inc()
				}
				return nil
			})
		}
		logx.Info("[CreateListener] 批量提交消息")
		//批量提交消息,无论消费是否异常
		sess.Commit()
		errGroup.Wait()
	}

	logx.Info("CreateEvtConsumer消费者停止运行")
	return nil
}

func convertMsg2ReadEvt(msg *sarama.ConsumerMessage) (domain.ReadEvent, error) {
	var event domain.ReadEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		logx.Errorf("[ReadEventListener] Failed to unmarshal message: %v", err)
		return domain.ReadEvent{}, err
	}
	return event, nil
}
