package logic

import (
	"context"
	"testing"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-code/internal/config"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-code/internal/repo/cache"
	"codexie.com/w-book-code/internal/repo/dao"
	"codexie.com/w-book-code/pkg/sms/provider"
	"codexie.com/w-book-common/ioc"

	"codexie.com/w-book-common/kafka/producer"
	"github.com/zeromicro/go-zero/core/conf"
)

func initCodeLogic() *CodeLogic {
	var c config.Config
	conf.MustLoad("/usr/local/go_project/w-book/app/code-service/etc/sms.yaml", &c)
	client := ioc.InitRedis(c.RedisConf)
	codeRedisCache := cache.NewCodeRedisCache(client)
	db := ioc.InitGormDB(c.MySQLConf)
	codeDao := dao.NewCodeDao(db)
	smsRepo := repo.NewSmsRepo(codeRedisCache, codeDao)
	saramaClient := ioc.InitKafkaClient(c.KafkaConf)
	producerProducer := producer.NewKafkaProducer(saramaClient)

	mem := provider.NewMemoryClient(c.SmsConf.Memory)
	tc := provider.NewTCSmsClient(c.SmsConf.TC)
	providerSmsLogic := NewProviderSmsLogic(mem, tc)
	prometheusSmsLogic := NewPrometheusSmsLogic(c.MetricConf, providerSmsLogic)
	aSyncSmsLogic := NewASyncSmsLogic(prometheusSmsLogic, smsRepo, producerProducer)
	codeLogic := NewCodeLogic(smsRepo, producerProducer, aSyncSmsLogic)
	return codeLogic
}

func TestSendCode(t *testing.T) {
	codeLogic := initCodeLogic()
	codeLogic.SendCode(context.Background(), &pb.SendCodeReq{Phone: "16602624578", Biz: "login"})
}

func TestVerifyCode(t *testing.T) {
	codeLogic := initCodeLogic()
	codeLogic.VerifyCode(context.Background(), &pb.VerifyCodeReq{Phone: "16602624578", Biz: "login", Code: "160521"})
	codeLogic.VerifyCode(context.Background(), &pb.VerifyCodeReq{Phone: "16602624578", Biz: "login", Code: "980963"})
}
