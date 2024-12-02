package logic

import (
	"context"

	"codexie.com/w-book-interact/api/internal/svc"
	"codexie.com/w-book-interact/api/pb/interact"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryInteractionInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryInteractionInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryInteractionInfoLogic {
	return &QueryInteractionInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryInteractionInfoLogic) QueryInteractionInfo(in *interact.QueryInteractionReq) (*interact.InteractionResult, error) {
	// todo: add your logic here and delete this line

	return &interact.InteractionResult{}, nil
}
