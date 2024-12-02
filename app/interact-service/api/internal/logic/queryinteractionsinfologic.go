package logic

import (
	"context"

	"codexie.com/w-book-interact/api/internal/svc"
	"codexie.com/w-book-interact/api/pb/interact"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryInteractionsInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryInteractionsInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryInteractionsInfoLogic {
	return &QueryInteractionsInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryInteractionsInfoLogic) QueryInteractionsInfo(in *interact.QueryInteractionsReq) (*interact.InteractionsInfo, error) {
	// todo: add your logic here and delete this line

	return &interact.InteractionsInfo{}, nil
}
