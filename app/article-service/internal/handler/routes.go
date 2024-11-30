package handler

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, articleHandler *ArticleHandler) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/article/edit",
				Handler: articleHandler.EditArticle,
			},
			{
				Method:  http.MethodPost,
				Path:    "/article/publish",
				Handler: articleHandler.Publish,
			},
			{
				Method:  http.MethodGet,
				Path:    "/article/view",
				Handler: articleHandler.ViewArticle,
			},
			{
				Method:  http.MethodGet,
				Path:    "/article/list",
				Handler: articleHandler.FindPage,
			},
			{
				Method:  http.MethodGet,
				Path:    "/article/toplike",
				Handler: articleHandler.TopLikeArticles,
			},
		},
		rest.WithPrefix("/v1"),
		rest.WithTimeout(3000*time.Millisecond),
		rest.WithMaxBytes(1048576000),
	)
}
