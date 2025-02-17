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

type CommentHandler struct {
	ctx          *svc.ServiceContext
	commentLogic *logic.CommentLogic
}

func NewCommentHandler(ctx *svc.ServiceContext, commentLogic *logic.CommentLogic) *CommentHandler {
	return &CommentHandler{
		ctx:          ctx,
		commentLogic: commentLogic,
	}
}

// 发表评论
func (h *CommentHandler) PublishComment(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.CommentReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	err := h.commentLogic.PublishComment(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(nil)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

// 分页查询热点资源的根评论
func (h *InteractHandler) RootCommentsPage(w http.ResponseWriter, r *http.Request) {
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

// 分页查询热点资源的子评论
func (h *InteractHandler) ChildCommentPage(w http.ResponseWriter, r *http.Request) {
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

// 点赞/取消点赞评论
func (h *InteractHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
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
