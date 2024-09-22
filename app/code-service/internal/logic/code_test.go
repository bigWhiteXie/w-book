package logic

import (
	"context"
	"testing"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-code/internal/config"
	"codexie.com/w-book-code/internal/kafka/producer"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-code/internal/repo/cache"
	"codexie.com/w-book-code/internal/repo/dao"
	"codexie.com/w-book-code/internal/svc"
	"codexie.com/w-book-code/pkg/sms"
	"github.com/zeromicro/go-zero/core/conf"
)

func initSvc() *svc.ServiceContext {
	var c config.Config
	conf.MustLoad("/usr/local/go_project/w-book/app/code-service/etc/sms.yaml", &c)
	ctx := svc.NewServiceContext(c)
	sms.InitSmsClient(c.SmsConf, ctx.Cache)
	return ctx
}

func TestSendCode(t *testing.T) {
	svc := initSvc()
	codeLogic := NewCodeLogic(repo.NewCodeRepo(cache.NewRedisCache(svc.Cache), dao.NewCodeDao(svc.DB)), producer.NewKafkaProducer(svc.KafkaProvider))
	codeLogic.SendCode(context.Background(), &pb.SendCodeReq{Phone: "16602624578", Biz: "login"})
}

func TestVerifyCode(t *testing.T) {
	svc := initSvc()
	codeLogic := NewCodeLogic(repo.NewCodeRepo(cache.NewRedisCache(svc.Cache), dao.NewCodeDao(svc.DB)), producer.NewKafkaProducer(svc.KafkaProvider))
	codeLogic.VerifyCode(context.Background(), &pb.VerifyCodeReq{Phone: "16602624578", Biz: "login", Code: "160521"})
	codeLogic.VerifyCode(context.Background(), &pb.VerifyCodeReq{Phone: "16602624578", Biz: "login", Code: "980963"})
}
