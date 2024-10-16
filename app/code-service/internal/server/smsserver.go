// Code generated by goctl. DO NOT EDIT.
// Source: sms.proto

package server

import (
	"context"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-code/internal/kafka/producer"
	"codexie.com/w-book-code/internal/repo"

	"codexie.com/w-book-code/internal/logic"
	"codexie.com/w-book-code/internal/svc"
)

type SMSServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedSMSServer
	codeLogic *logic.CodeLogic
}

func NewSMSServer(svcCtx *svc.ServiceContext, smsRepo repo.SmsRepo, kafkaProvider *producer.KafkaProducer) *SMSServer {
	return &SMSServer{
		svcCtx:    svcCtx,
		codeLogic: logic.NewCodeLogic(smsRepo,kafkaProvider),
	}
}

func (s *SMSServer) SendCode(ctx context.Context, in *pb.SendCodeReq) (*pb.SendCodeResp, error) {
	return s.codeLogic.SendCode(ctx, in)
}

func (s *SMSServer) VerifyCode(ctx context.Context, in *pb.VerifyCodeReq) (*pb.VerifyCodeResp, error) {
	return s.codeLogic.VerifyCode(ctx, in)
}
