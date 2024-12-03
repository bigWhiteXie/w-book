package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"codexie.com/w-book-common/metric"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

type BatchConsumer[T any] struct {
	handler   func(objs []T, msgs []*sarama.ConsumerMessage) error
	client    sarama.ConsumerGroup
	topic     string
	group     string
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	batchSize int
}

func NewBatchConsumer[T any](topic string, client sarama.Client, group string, batchSize int, handler func(objs []T, msgs []*sarama.ConsumerMessage) error) *BatchConsumer[T] {
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V2_1_0_0
	consumerClient, err := sarama.NewConsumerGroupFromClient(group, client)
	if err != nil {
		panic(fmt.Sprintf("unable to create kafka consumer group, cause:%s", err))
	}
	if err != nil {
		panic(err)
	}
	return &BatchConsumer[T]{
		client:    consumerClient,
		topic:     topic,
		group:     group,
		batchSize: batchSize,
		handler:   handler,
	}
}
func (s *BatchConsumer[T]) StartListner() {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		s.cancel = cancel
		// group关闭后consume方法会自动退出并释放资源
		err := s.client.Consume(ctx, []string{s.topic}, s)
		if err != nil {
			logx.Errorf("[BatchConsumer] fail to consumer kafka msg,cause:%s", err)
		}
	}()
}

func (s *BatchConsumer[T]) Stop() {
	//s.client.Consume方法会尝试停止从Kafka主题中拉取消息，并关闭与Kafka broker的连接
	s.cancel()
	// 等待所有消费者goroutine完成
	s.wg.Wait()
	// 关闭Kafka客户端
	if err := s.client.Close(); err != nil {
		logx.Errorf("[BatchConsumer] fail to close kafka client,cause:%s", err)
	}
}
func (BatchConsumer[T]) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (BatchConsumer[T]) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h BatchConsumer[T]) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var (
		msgs []*sarama.ConsumerMessage
		objs []T
	)
	for {
		msgs = make([]*sarama.ConsumerMessage, 0, h.batchSize)
		objs = make([]T, 0, h.batchSize)
		ctx, cancelFun := context.WithTimeout(context.Background(), time.Second)
		timeReached := false
		defer cancelFun()
		for i := 0; i < h.batchSize && !timeReached; i++ {
			select {
			case <-ctx.Done():
				timeReached = true
			case msg, ok := <-claim.Messages():
				if !ok {
					return nil
				}
				msgs = append(msgs, msg)
				var obj T
				if err := json.Unmarshal(msg.Value, &obj); err != nil {
					logx.Errorf("[BatchConsumer] fail to unmarshal msg,cause:%s", err)
					break
				}
				objs = append(objs, obj)
				sess.MarkMessage(msg, "")
				logx.Infof("添加msg:%s", msg.Value)
			}
		}

		//批量消费消息
		// todo: 埋点prometheus
		if len(msgs) > 0 {
			metric.ConsumeCountCounter.WithLabelValues(h.topic, h.group).Add(float64(len(msgs)))
			start := time.Now()
			if err := h.handler(objs, msgs); err != nil {
				metric.ConsumeErrCounter.WithLabelValues(h.topic, h.group, err.Error()).Add(float64(len(msgs)))
			}
			metric.ConsumeTimeHistogram.WithLabelValues(h.topic, h.group).Observe(float64(time.Since(start).Milliseconds()))
		}
	}
}
