package logic

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-common/common/codeerr"
	"codexie.com/w-book-common/producer"

	"codexie.com/w-book-code/api/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type CodeLogic struct {
	codeRepo      repo.SmsRepo
	kafkaProvider producer.Producer
	smsService    SmsService
	logx.Logger
}

func NewCodeLogic(repo repo.SmsRepo, provider producer.Producer, smsService *ASyncSmsLogic) *CodeLogic {
	return &CodeLogic{
		codeRepo:      repo,
		smsService:    smsService,
		kafkaProvider: provider,
	}
}

func (l *CodeLogic) SendCode(ctx context.Context, in *pb.SendCodeReq) (*pb.SendCodeResp, error) {
	//1.生成验证码
	randomCode := generateVerificationCode()
	log := logx.WithContext(ctx)
	//2.执行lua脚本，校验并缓存验证码
	k := key(in.Biz, in.Phone)
	err := l.codeRepo.StoreCode(ctx, k, randomCode, sendCodeLuaTemplate())
	if err != nil {
		log.Errorf("[SendCode] %v", err)
		return nil, codeerr.ToGrpcErr(err)
	}

	//调用短信服务发送短信
	args := map[string]string{"code": randomCode}
	if err = l.smsService.SendSms(ctx, in.Phone, args); err != nil {
		return nil, err
	}

	return &pb.SendCodeResp{Result: pb.Success()}, nil
}

func (l *CodeLogic) VerifyCode(ctx context.Context, in *pb.VerifyCodeReq) (*pb.VerifyCodeResp, error) {
	// 1. 判断校验次数是否超过限制
	if err := l.codeRepo.VerifyCode(ctx, key(in.Biz, in.Phone), in.Code, verifyCodeLuaScript()); err != nil {
		logx.Errorf("[VerifyCode] %v", err)
		return nil, codeerr.ToGrpcErr(err)
	}

	return &pb.VerifyCodeResp{}, nil
}

func generateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())   // 设置随机种子
	code := rand.Intn(900000) + 100000 // 生成六位随机数
	return fmt.Sprintf("%06d", code)   // 格式化为六位数字
}

func key(buz, phone string) string {
	return fmt.Sprintf("%s:code:%s", buz, phone)
}
