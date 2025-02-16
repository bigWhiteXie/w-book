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
		req  types.AddCommentReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	cmt, err := h.commentLogic.AddComment(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(cmt)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

// 分页查询热点资源的根评论
func (h *CommentHandler) RootCommentsPage(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.GetRootCommentsReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	cmts, err := h.commentLogic.GetRootComments(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(cmts)
	}

	httpx.OkJsonCtx(r.Context(), w, resp)
}

// 分页查询热点资源的子评论
func (h *CommentHandler) ChildCommentPage(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.GetChildCommentsReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	cmts, err := h.commentLogic.GetChildComments(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(cmts)
	}

	httpx.OkJsonCtx(r.Context(), w, resp)
}

// 点赞/取消点赞评论
func (h *CommentHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.LikeCommentReq
		resp *response.Response
	)
	if err := httpx.ParsePath(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	err := h.commentLogic.LikeComment(r.Context(), req.CommentID, req.IsLike)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(nil)
	}

	httpx.OkJsonCtx(r.Context(), w, resp)
}
