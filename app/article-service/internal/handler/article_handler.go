package handler

import (
	"net/http"

	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/svc"
	"codexie.com/w-book-article/internal/types"
	"codexie.com/w-book-common/common/codeerr"
	"codexie.com/w-book-common/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type ArticleHandler struct {
	ctx          *svc.ServiceContext
	articleLogic *logic.ArticleLogic
}

func NewArticleHandler(ctx *svc.ServiceContext, articleLogic *logic.ArticleLogic) *ArticleHandler {
	return &ArticleHandler{
		ctx:          ctx,
		articleLogic: articleLogic,
	}
}

func (h *ArticleHandler) EditArticle(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.EditArticleReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	id, err := h.articleLogic.Edit(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(id)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *ArticleHandler) Publish(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.EditArticleReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	id, err := h.articleLogic.Publish(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(id)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *ArticleHandler) FindPage(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.ArticlePageReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	artList, err := h.articleLogic.Page(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(artList)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *ArticleHandler) ViewArticle(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.ArticleViewReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	article, err := h.articleLogic.ViewArticle(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(article)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}
