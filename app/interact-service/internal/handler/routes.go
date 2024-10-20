package handler

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, interactHandler *InteractHandler) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/resource/like",
				Handler: interactHandler.LikeResource,
			},
			{
				Method:  http.MethodPost,
				Path:    "/resource/collection",
				Handler: interactHandler.OperateCollection,
			},
			{
				Method:  http.MethodPost,
				Path:    "/resource/collect",
				Handler: interactHandler.OperateCollectionItem,
			},
		},
		rest.WithPrefix("/v1"),
		rest.WithTimeout(3000*time.Millisecond),
		rest.WithMaxBytes(1048576000),
	)
}
