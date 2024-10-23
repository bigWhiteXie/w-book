package server

import (
	"context"

	"codexie.com/w-book-interact/api/pb/interact"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/svc"
	"codexie.com/w-book-interact/internal/types"
)

type InteractionServer struct {
	svcCtx *svc.ServiceContext
	interact.UnimplementedInteractionServer
	logic *logic.InteractLogic
}

func NewInteractionServer(svcCtx *svc.ServiceContext, interactLogic *logic.InteractLogic) *InteractionServer {
	return &InteractionServer{
		svcCtx: svcCtx,
		logic:  interactLogic,
	}
}

func (s *InteractionServer) QueryInteractionInfo(ctx context.Context, in *interact.QueryInteractionReq) (*interact.InteractionResult, error) {
	stat, err := s.logic.QueryStatInfo(ctx, &types.LikeResourceReq{
		Biz:   in.Biz,
		BizId: in.BizId,
	})

	if err != nil {
		return nil, err
	}
	return &interact.InteractionResult{
		ReadCnt:    stat.ReadCnt,
		LikeCnt:    stat.LikeCnt,
		CollectCnt: stat.CollectCnt,
	}, err
}

func (s *InteractionServer) IncreReadCnt(ctx context.Context, in *interact.AddReadCntReq) (*interact.CommonResult, error) {
	err := s.logic.AddRead(ctx, in.Biz, in.BizId)
	if err != nil {
		return nil, err
	}
	return &interact.CommonResult{
		Msg: "ok",
	}, nil
}
