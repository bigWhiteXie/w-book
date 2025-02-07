package server

import (
	"context"

	"codexie.com/w-book-account/api/pb"
	"codexie.com/w-book-account/internal/service"
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

func (s *AccountServer) Credit(ctx context.Context, req *pb.CreditRequest) (*pb.CreditResponse, error) {
	return s.svc.Credit(ctx, req)
}
