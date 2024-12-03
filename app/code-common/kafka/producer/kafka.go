package producer

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type KafkaProducer struct {
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
}

func NewKafkaProducer(client sarama.Client) Producer {
	sc, err := sarama.NewSyncProducerFromClient(client)
	asc, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		panic("fail to init kafka producer,cause:" + err.Error())
	}
	return &KafkaProducer{syncProducer: sc, asyncProducer: asc}
}

func (p *KafkaProducer) SendSync(ctx context.Context, topic string, msg string, opts ...ProducerOption) error {
	logger := logx.WithContext(ctx)

	message, err := p.buildMessage(ctx, topic, msg, opts...)
	if err != nil {
		return errors.Wrapf(err, "[KafkaProducer] 构建消息[%s]失败:%s", msg, err)
	}
	partition, offset, err := p.syncProducer.SendMessage(message)
	if err != nil {
		return errors.Wrapf(err, "[KafkaProducer] 投递消息[%s]失败:%s", msg, err)
	}
	logger.Infof("[kafka producer] 发送成功, topic:%s, partition:%d, offset: %d", topic, partition, offset)
	return nil
}

func (p *KafkaProducer) SendAsync(ctx context.Context, topic string, msg string, onError func(error), opts ...ProducerOption) {
	message, err := p.buildMessage(ctx, topic, msg, opts...)
	if err != nil {
		onError(err)
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.WithContext(ctx).Errorf("[KafkaProducer] 异步发送消息时发生panic: %v", r)
				if onError != nil {
					onError(fmt.Errorf("panic: %v", r))
				}
			}
		}()
		partition, offset, err := p.syncProducer.SendMessage(message)
		if err != nil {
			onError(err)
			return
		}
		logx.Infof("[kafka producer] 发送成功, topic:%s, partition:%d, offset: %d", topic, partition, offset)
	}()
}

func (p *KafkaProducer) buildMessage(ctx context.Context, topic string, msg string, opts ...ProducerOption) (*sarama.ProducerMessage, error) {
	if topic == "" || msg == "" {
		return nil, errors.New("topic 和 msg 是必选参数")
	}
	options := ApplyOptions(opts...)
	logx.Infof("[kafka producer]发送消息: topic=%s, msg=%s, key=%s, headers=%v\n", topic, msg, options.Key, options.Headers)
	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msg),
	}
	if len(options.Headers) != 0 {
		headerRes := make([]sarama.RecordHeader, 0, len(options.Headers))
		for k, v := range options.Headers {
			headerRes = append(headerRes, sarama.RecordHeader{
				Key:   []byte(k),
				Value: []byte(v),
			})
		}
		message.Headers = headerRes
	}

	if options.Key != "" {
		message.Key = sarama.StringEncoder(options.Key)
	}
	return message, nil
}
