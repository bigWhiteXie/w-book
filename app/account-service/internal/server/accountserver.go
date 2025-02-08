package server

import (
	"context"

	"codexie.com/w-book-account/api/pb"
	"codexie.com/w-book-account/internal/service"
	"github.com/zeromicro/go-zero/core/logx"
)

type AccountServer struct {
	pb.UnimplementedAccountServer
	svc *service.AccountService
}

func NewAccountServer(svc *service.AccountService) *AccountServer {
	return &AccountServer{
		svc: svc,
	}
}

func (s *AccountServer) Credit(ctx context.Context, req *pb.CreditRequest) (resp *pb.CreditResponse, err error) {
	if resp, err = s.svc.Credit(ctx, req); err != nil {
		logx.Errorw("分账失败", logx.Field("req", req), logx.Field("err", err))
		return nil, err
	}
	return &pb.CreditResponse{
		Code:    0,
		Message: "Success",
	}, nil
}
