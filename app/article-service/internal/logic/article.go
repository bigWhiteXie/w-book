package logic

import (
	"context"

	"codexie.com/w-book-article/internal/types"
)

type ArticleLogic struct {
}

func NewArticleLogic() *ArticleLogic {
	return &ArticleLogic{}
}

func (l *ArticleLogic) Edit(ctx context.Context, req *types.EditArticleReq) (resp *types.EditArticleResp, err error) {
	// todo: add your logic here and delete this line

	return
}
