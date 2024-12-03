package handler

import (
	"net/http"

	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-common/response"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/svc"
	"codexie.com/w-book-interact/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type InteractHandler struct {
	ctx           *svc.ServiceContext
	interactLogic *logic.InteractLogic
}

func NewInteractHandler(ctx *svc.ServiceContext, interactLogic *logic.InteractLogic) *InteractHandler {
	return &InteractHandler{
		ctx:           ctx,
		interactLogic: interactLogic,
	}
}

func (h *InteractHandler) LikeResource(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.OpResourceReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	err := h.interactLogic.Like(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(nil)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *InteractHandler) OperateCollection(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.CollectionReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	err := h.interactLogic.AddOrDelCollection(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(nil)
	}

	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *InteractHandler) OperateCollectionItem(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.CollectResourceReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	err := h.interactLogic.Collect(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(nil)
	}

	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *InteractHandler) TopLikeResource(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.TopLikeReq
		resp *response.Response
		res  []int64
	)
	if err := httpx.ParsePath(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	res, err := h.interactLogic.GetTopLike(r.Context(), req.Biz)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(res)
	}

	httpx.OkJsonCtx(r.Context(), w, resp)
}
