package event

import (
	"codexie.com/w-book-common/kafka/consumer"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"github.com/IBM/sarama"
)

func NewBatchReadEventListener(client sarama.Client, interactRepo repo.IInteractRepo) *consumer.BatchConsumer[domain.ReadEvent] {
	return consumer.NewBatchConsumer[domain.ReadEvent](domain.ReadEvtTopic, client, "interact-group", 100, interactRepo.HandleBatchReadV2)
}
