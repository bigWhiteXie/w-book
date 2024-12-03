package ioc

import (
	"github.com/IBM/sarama"
)

type KafkaConf struct {
	Brokers []string `json:"brokers"`
}

func InitKafkaClient(kafkaConf KafkaConf) sarama.Client {
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V2_1_0_0
	client, err := sarama.NewClient(kafkaConf.Brokers, saramaConf)
	if err != nil {
		panic(err)
	}
	return client
}
