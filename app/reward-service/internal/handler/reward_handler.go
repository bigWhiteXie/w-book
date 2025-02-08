package handler

import (
	"net/http"

	"codexie.com/w-book-common/user"
	"codexie.com/w-book-payment/pkg/constant"
	"codexie.com/w-book-reward/internal/domain"
	"codexie.com/w-book-reward/internal/logic"
	"codexie.com/w-book-reward/internal/svc"
	"codexie.com/w-book-reward/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type RewardHandler struct {
	ctx         *svc.ServiceContext
	rewardLogic *logic.RewardLogic
}

func NewRewardHandler(ctx *svc.ServiceContext, rewardLogic *logic.RewardLogic) *RewardHandler {
	return &RewardHandler{
		ctx:         ctx,
		rewardLogic: rewardLogic,
	}
}

func (h *RewardHandler) Reward(w http.ResponseWriter, r *http.Request) {
	var (
		req types.RewardReq
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	uid := user.GetUidByCtx(r.Context())
	rewardDomain := &domain.RewardRecord{
		Id:         int64(req.Id),
		Uid:        uid,
		AuthorId:   req.AuthorId,
		ResourceId: req.ResourceId,
		Biz:        req.Biz,
		Platform:   req.Platform,
		Amt:        req.Amt,
		Status:     constant.InitPayStatus,
	}

	resp, err := h.rewardLogic.RewardPcWEB(r.Context(), rewardDomain)
	if err != nil {
		httpx.Error(w, err)
		return
	}

	httpx.OkJson(w, resp)
}
