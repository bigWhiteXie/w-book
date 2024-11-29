package logic

import (
	"context"
	"testing"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-code/internal/config"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-code/internal/repo/cache"
	"codexie.com/w-book-code/internal/repo/dao"
	"codexie.com/w-book-code/internal/svc"
	"codexie.com/w-book-code/pkg/sms/provider"
	"codexie.com/w-book-common/producer"
	"github.com/zeromicro/go-zero/core/conf"
)

func initCodeLogic() *CodeLogic {
	var c config.Config
	conf.MustLoad("/usr/local/go_project/w-book/app/code-service/etc/sms.yaml", &c)
	svc := svc.NewServiceContext(c)
	repo := repo.NewSmsRepo(cache.NewRedisCache(svc.Cache), dao.NewCodeDao(svc.DB))
	KafkaProvider := producer.NewKafkaProducer(svc.KafkaProvider)
	mem := provider.NewMemoryClient(c.SmsConf.Memory)
	tc := provider.NewTCSmsClient(c.SmsConf.TC)
	providerSmsService := NewProviderSmsLogic(mem, tc)
	asyncSmsLogic := NewASyncSmsLogic(providerSmsService, repo)
	return NewCodeLogic(repo, KafkaProvider, asyncSmsLogic)
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
