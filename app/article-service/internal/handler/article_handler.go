package handler

import (
	"net/http"

	"codexie.com/w-book-article/internal/domain"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/svc"
	"codexie.com/w-book-article/internal/types"

	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-common/response"
	"codexie.com/w-book-common/user"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type ArticleHandler struct {
	ctx          *svc.ServiceContext
	articleLogic *logic.ArticleLogic
	rankingLogic *logic.RankingLogic
}

func NewArticleHandler(ctx *svc.ServiceContext, articleLogic *logic.ArticleLogic, rankingLogic *logic.RankingLogic) *ArticleHandler {
	return &ArticleHandler{
		ctx:          ctx,
		articleLogic: articleLogic,
		rankingLogic: rankingLogic,
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
	id := user.GetUidByCtx(r.Context())
	artDomain := &domain.Article{
		Id:      int64(req.Id),
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: int64(id),
		},
	}
	id, err := h.articleLogic.Edit(r.Context(), artDomain)
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
	uid := user.GetUidByCtx(r.Context())
	artDomain := &domain.Article{
		Id:      int64(req.Id),
		Title:   req.Title,
		Content: req.Content,
		Status:  domain.ArticlePublishedStatus,
		Author: domain.Author{
			Id: uid,
		},
	}
	id, err := h.articleLogic.Publish(r.Context(), artDomain)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(id)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *ArticleHandler) TopLikeArticles(w http.ResponseWriter, r *http.Request) {
	var (
		resp *response.Response
	)

	articles, err := h.rankingLogic.GetTopArticles(r.Context())
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(articles)
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
	id := user.GetUidByCtx(r.Context())
	artList, err := h.articleLogic.Page(r.Context(), id, req.Page, req.Size)
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
	if err := httpx.ParseForm(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	article, err := h.articleLogic.ViewArticle(r.Context(), req.Id, req.Published > 0)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(article)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}
