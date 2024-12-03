package ioc

import (
	"codexie.com/w-book-code/internal/logic"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-common/kafka/producer"
)

func InitCodeLogic(repo *repo.SmsRepo, provider producer.Producer, smsService *logic.ASyncSmsLogic) *logic.CodeLogic {
	return logic.NewCodeLogic(repo, provider, smsService)
}
