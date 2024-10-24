package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	batchSize         = 100
	batchTimeout      = 2 * time.Second
	readConsumerGroup = "read-evt-consumer-group"
)

// ：kafka-topics.sh --bootstrap-server 192.168.126.100:9092 --topic read-evt-topic --create --partitions 1 --replication-factor 1
type ReadEventListener struct {
	topic        string
	client       sarama.ConsumerGroup
	interactRepo repo.IInteractRepo
	once         sync.Once
}

// 用户针对需求对consumerGroup做好配置
// 短信服务依赖该consumerGroup启动消费者
func NewReadEventListener(config config.Config, interactRepo repo.IInteractRepo) *ReadEventListener {
	conf := config.KafkaConf
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V2_1_0_0
	client, err := sarama.NewConsumerGroup(conf.Brokers, readConsumerGroup, saramaConf)
	if err != nil {
		panic(fmt.Sprintf("unable to create kafka consumer group, cause:%s", err))
	}

	return &ReadEventListener{
		client:       client,
		interactRepo: interactRepo,
	}
}

func (s *ReadEventListener) StartListner() {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// group关闭后consume方法会自动退出并释放资源
		err := s.client.Consume(ctx, []string{domain.ReadEvtTopic}, s)
		if err != nil {
			logx.Errorf("[ReadEventListener] fail to kafka msg,cause:%s", err)
		}
	}()
}

func (s *ReadEventListener) Stop() {
	s.once.Do(func() {
		s.client.Close()
	})
}
func (ReadEventListener) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (ReadEventListener) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h ReadEventListener) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		msgBatch := make([]*sarama.ConsumerMessage, 0, batchSize)
		done := false
		ctx, cancelFun := context.WithTimeout(context.Background(), batchTimeout)
		defer cancelFun()
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-claim.Messages():
				if !ok {
					cancelFun()
					//通道关闭，处理完剩余的事件
					if len(msgBatch) > 0 {
						if evtBatch, err := convertMsgs2ReadEvts(msgBatch); err == nil {
							err = h.interactRepo.HandleBatchRead(context.Background(), evtBatch)
							if err != nil {
								logx.Errorf("[ReadEventListner] 批量处理浏览事件异常:%s", err)
							}
						} else {
							logx.Errorf("[ReadEvtListener] 反序列化消息异常:%s", err)
						}
					}
					return nil
				}
				sess.MarkMessage(msg, "")
				msgBatch = append(msgBatch, msg)
				logx.Infof("添加msg:%s", msg.Value)
			}
		}
		cancelFun()
		if len(msgBatch) > 0 {
			//批次结束时不管消费成功与否都提交消息
			sess.Commit()
			if evtBatch, err := convertMsgs2ReadEvts(msgBatch); err == nil {
				err = h.interactRepo.HandleBatchRead(context.Background(), evtBatch)
				if err != nil {
					logx.Errorf("[ReadEventListner] 批量处理浏览事件异常:%s", err)
				}
			} else {
				logx.Errorf("[ReadEvtListener] 反序列化消息异常:%s", err)
			}
		}
	}
}

func convertMsgs2ReadEvts(msgs []*sarama.ConsumerMessage) ([]domain.ReadEvent, error) {
	evtBatch := make([]domain.ReadEvent, 0, len(msgs))
	for _, msg := range msgs {
		var event domain.ReadEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			logx.Errorf("[ReadEventListener] Failed to unmarshal message: %v", err)
			return nil, err
		}
		evtBatch = append(evtBatch, event)
	}
	return evtBatch, nil
}
