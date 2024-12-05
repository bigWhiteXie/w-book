package startup

import (
	"github.com/IBM/sarama"
)

type KafkaConf struct {
	Brokers []string `json:"brokers"`
}

func InitKafkaClient() sarama.Client {
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V2_1_0_0
	saramaConf.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"127.0.0.1:19092"}, saramaConf)
	if err != nil {
		panic(err)
	}
	return client
}
