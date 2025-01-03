// Code generated by goctl. DO NOT EDIT.
// Source: sms.proto

package server

import (
	"context"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-code/internal/repo"
	"codexie.com/w-book-common/producer"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"codexie.com/w-book-code/internal/logic"
	"codexie.com/w-book-code/internal/svc"
)

type SMSServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedSMSServer
	codeLogic *logic.CodeLogic
}

func NewSMSServer(svcCtx *svc.ServiceContext, smsRepo repo.SmsRepo, producer producer.Producer, smsService *logic.ASyncSmsLogic) *SMSServer {
	return &SMSServer{
		svcCtx:    svcCtx,
		codeLogic: logic.NewCodeLogic(smsRepo,producer,smsService),
	}
}

func (s *SMSServer) SendCode(ctx context.Context, in *pb.SendCodeReq) (resp *pb.SendCodeResp, err error) {
	if resp, err = s.codeLogic.SendCode(ctx, in); err != nil {
		logx.Errorf("[SMSServer_SendCode] 发送短信失败,cause=%s, stack:%v", errors.Cause(err),err)
		return nil, err
	}

	return resp, nil
}

func (s *SMSServer) VerifyCode(ctx context.Context, in *pb.VerifyCodeReq) (*pb.VerifyCodeResp, error) {
	return s.codeLogic.VerifyCode(ctx, in)
}
