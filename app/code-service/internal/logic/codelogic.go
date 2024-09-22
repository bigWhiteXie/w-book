package logic

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"codexie.com/w-book-code/internal/kafka/producer"
	"codexie.com/w-book-code/internal/model"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-user/pkg/common/codeerr"

	"codexie.com/w-book-code/api/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

var topic = "sms-topic"

type CodeLogic struct {
	codeRepo      repo.CodeRepo
	kafkaProvider *producer.KafkaProducer
	logx.Logger
}

func NewCodeLogic(repo repo.CodeRepo, provider *producer.KafkaProducer) *CodeLogic {
	return &CodeLogic{
		codeRepo:      repo,
		kafkaProvider: provider,
	}
}

func (l *CodeLogic) SendCode(ctx context.Context, in *pb.SendCodeReq) (*pb.SendCodeResp, error) {
	//1.生成验证码
	randomCode := generateVerificationCode()

	//2.执行lua脚本，校验并缓存验证码
	k := key(in.Biz, in.Phone)
	err := l.codeRepo.StoreCode(ctx, k, randomCode, sendCodeLuaTemplate())
	if err != nil {
		logx.Errorf("[send code] %v", err)
		return nil, codeerr.ToGrpcErr(err)
	}

	//将短信内容记录到数据库中
	record := model.NewSmsRecord(in.Phone, map[string]string{"code": randomCode})
	if err = l.codeRepo.SaveSmsRecord(ctx, record); err != nil {
		logx.Errorf("[send code] %v", err)
		return nil, err
	}

	l.kafkaProvider.Send(ctx, topic, strconv.Itoa(record.ID), func(err error) {
		logx.Errorf("fail to send msg to kafka, cause:%s", err)
	})

	return &pb.SendCodeResp{Result: pb.Success()}, nil
}

func (l *CodeLogic) VerifyCode(ctx context.Context, in *pb.VerifyCodeReq) (*pb.VerifyCodeResp, error) {
	// 1. 判断校验次数是否超过限制
	if err := l.codeRepo.VerifyCode(ctx, key(in.Biz, in.Phone), in.Code, verifyCodeLuaScript()); err != nil {
		logx.Errorf("[verify code] %v", err)
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
