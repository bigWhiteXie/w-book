package ioc

import (
	"codexie.com/w-book-code/internal/logic"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-code/pkg/sms"
	"codexie.com/w-book-code/pkg/sms/provider"
	"codexie.com/w-book-common/kafka/producer"
	"codexie.com/w-book-common/metric"
)

func InitPrometheusSmsService(conf sms.SmsConf, labelConf metric.ConstMetricLabelsConf) *logic.PrometheusSmsLogic {
	mem := provider.NewMemoryClient(conf.Memory)
	tc := provider.NewTCSmsClient(conf.TC)
	providerSmsLogic := logic.NewProviderSmsLogic(mem, tc)
	return logic.NewPrometheusSmsLogic(labelConf, providerSmsLogic)
}

func InitKafkaSmsService(prometheusSmsService *logic.PrometheusSmsLogic, smsRepo *repo.SmsRepo, kafkaProducer producer.Producer) *logic.ASyncSmsLogic {
	return logic.NewASyncSmsLogic(prometheusSmsService, smsRepo, kafkaProducer)
}
