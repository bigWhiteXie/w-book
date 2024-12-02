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
	stat, err := s.logic.QueryStatInfo(ctx, &types.OpResourceReq{
		Biz:   in.Biz,
		BizId: in.BizId,
		Uid:   in.Uid,
	})

	if err != nil {
		return nil, err
	}
	return &interact.InteractionResult{
		ReadCnt:     stat.ReadCnt,
		LikeCnt:     stat.LikeCnt,
		CollectCnt:  stat.CollectCnt,
		IsLiked:     stat.IsLiked,
		IsCollected: stat.IsCollected,
	}, err
}

func (s *InteractionServer) QueryInteractionsInfo(ctx context.Context, in *interact.QueryInteractionsReq) (*interact.InteractionsResp, error) {
	stats, err := s.logic.QueryInteractionInfos(ctx, in.Biz, in.BizIds)

	if err != nil {
		return nil, err
	}
	res := make([]*interact.InteractionResult, 0, len(stats))
	for _, stat := range stats {
		res = append(res, &interact.InteractionResult{
			BizId:       stat.BizId,
			ReadCnt:     stat.ReadCnt,
			LikeCnt:     stat.LikeCnt,
			CollectCnt:  stat.CollectCnt,
			IsLiked:     stat.IsLiked,
			IsCollected: stat.IsCollected,
		})
	}

	return &interact.InteractionsResp{
		Interactions: res,
	}, nil
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

func (s *InteractionServer) TopLike(ctx context.Context, in *interact.TopLikeReq) (*interact.TopLikeResp, error) {
	ids, err := s.logic.GetTopLike(ctx, in.Biz)
	if err != nil {
		return nil, err
	}
	return &interact.TopLikeResp{
		Items: ids,
	}, nil
}
