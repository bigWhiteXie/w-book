package event

import (
	"codexie.com/w-book-common/kafka/consumer"
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
)

func NewBatchReadEventListener(config config.Config, interactRepo repo.IInteractRepo) *consumer.BatchConsumer[domain.ReadEvent] {
	conf := config.KafkaConf
	return consumer.NewBatchConsumer[domain.ReadEvent](domain.ReadEvtTopic, conf.Brokers, "interact-group", 100, interactRepo.HandleBatchReadV2)
}
