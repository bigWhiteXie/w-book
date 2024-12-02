package logic

import (
	"context"

	"codexie.com/w-book-interact/api/internal/svc"
	"codexie.com/w-book-interact/api/pb/interact"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncreReadCntLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncreReadCntLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncreReadCntLogic {
	return &IncreReadCntLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IncreReadCntLogic) IncreReadCnt(in *interact.AddReadCntReq) (*interact.CommonResult, error) {
	// todo: add your logic here and delete this line

	return &interact.CommonResult{}, nil
}
