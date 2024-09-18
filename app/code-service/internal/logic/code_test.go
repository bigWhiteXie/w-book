package logic

import (
	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-code/internal/config"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-code/internal/svc"
	"codexie.com/w-book-code/pkg/sms"
	"context"
	"github.com/zeromicro/go-zero/core/conf"
	"testing"
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
	codeLogic := NewCodeLogic(repo.NewRedisCache(svc.Cache))
	codeLogic.SendCode(context.Background(), &pb.SendCodeReq{Phone: "16602624578", Biz: "login"})
}

func TestVerifyCode(t *testing.T) {
	svc := initSvc()
	codeLogic := NewCodeLogic(repo.NewRedisCache(svc.Cache))
	codeLogic.VerifyCode(context.Background(), &pb.VerifyCodeReq{Phone: "16602624578", Biz: "login", Code: "160521"})
	codeLogic.VerifyCode(context.Background(), &pb.VerifyCodeReq{Phone: "16602624578", Biz: "login", Code: "980963"})
}
