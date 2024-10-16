package handler

import (
	"net/http"

	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/svc"
	"codexie.com/w-book-article/internal/types"
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

	var req types.EditArticleReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	resp, err := h.articleLogic.Edit(r.Context(), &req)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
	} else {
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
