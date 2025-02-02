package server

import (
	"context"

	"codexie.com/w-book-payment/api/pb"
	"codexie.com/w-book-payment/internal/svc"
	"codexie.com/w-book-payment/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
)

type PaymentServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedPaymentServer
}

func NewPaymentServer(svcCtx *svc.ServiceContext) *PaymentServer {
	return &PaymentServer{
		svcCtx: svcCtx,
	}
}

func (s *PaymentServer) PrePay(ctx context.Context, in *pb.PrepayReq) (*pb.PrepayResp, error) {
	payLogic, err := s.svcCtx.GetPayLogic(constant.Platform(in.Platform))
	if err != nil {
		logx.WithContext(ctx).Errorf("no platform found for %s", in.Platform)
		return nil, err
	}
	return payLogic.PrePay(ctx, in)
}

func (s *PaymentServer) GetPayment(ctx context.Context, in *pb.QueryPaymentReq) (*pb.QueryPaymentResp, error) {
	return nil, nil
}
