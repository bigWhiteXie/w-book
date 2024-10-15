package producer

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

type KafkaProducer struct {
	producer sarama.AsyncProducer
}

func NewKafkaProducer(producer sarama.AsyncProducer) *KafkaProducer {
	return &KafkaProducer{producer: producer}
}

func (p *KafkaProducer) Send(ctx context.Context, topic string, msg string, errFunc func(err error)) {
	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msg),
	}
	p.producer.Input() <- message

	go func() {
		select {
		case success := <-p.producer.Successes():
			logx.WithContext(ctx).Infof("Message sent to partition %d at offset %d\n", success.Partition, success.Offset)
		case err := <-p.producer.Errors():
			errFunc(err)
		}
	}()
}
