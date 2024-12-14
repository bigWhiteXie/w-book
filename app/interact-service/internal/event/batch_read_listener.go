package event

import (
	"context"

	"codexie.com/w-book-common/kafka/consumer"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

type ReadEvtListener struct {
	interactRepo  repo.IInteractRepo
	batchConsumer *consumer.BatchConsumer[domain.ReadEvent]
}

func NewBatchReadEventListener(client sarama.Client, interactRepo repo.IInteractRepo) *ReadEvtListener {
	l := &ReadEvtListener{
		interactRepo: interactRepo,
	}
	l.batchConsumer = consumer.NewBatchConsumer[domain.ReadEvent](domain.ReadEvtTopic, client, "interact-group", 100, l.handleBatchRead)

	return l
}

func (c *ReadEvtListener) handleBatchRead(eventBatch []domain.ReadEvent, msgs []*sarama.ConsumerMessage) error {
	logx.WithContext(context.Background()).Infof("批量阅读事件,长度:%d", len(eventBatch))
	err := c.interactRepo.HandleBatchRead(context.Background(), eventBatch)
	if err != nil {
		logx.Errorf("处理批量阅读事件失败,原因:%s", err)
	}
	return err
}

func (c *ReadEvtListener) Start() {
	c.batchConsumer.StartListner()
}

func (c *ReadEvtListener) Stop() {
	c.batchConsumer.Stop()
}
