package logic

import (
	"context"

	"codexie.com/w-book-interact/api/internal/svc"
	"codexie.com/w-book-interact/api/pb/interact"

	"github.com/zeromicro/go-zero/core/logx"
)

type TopLikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTopLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TopLikeLogic {
	return &TopLikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TopLikeLogic) TopLike(in *interact.TopLikeReq) (*interact.TopLikeResp, error) {
	// todo: add your logic here and delete this line

	return &interact.TopLikeResp{}, nil
}
