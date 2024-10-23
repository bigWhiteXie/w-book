package logic

import (
	"context"
	"encoding/json"

	"codexie.com/w-book-article/internal/domain"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-article/internal/types"
	"codexie.com/w-book-common/producer"
	"codexie.com/w-book-common/user"
	"codexie.com/w-book-interact/api/pb/interact"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleLogic struct {
	authorRepo  repo.IAuthorRepository
	readerRepo  repo.IReaderRepository
	interactRpc interact.InteractionClient
	producer    producer.Producer
}

func NewArticleLogic(authorRepo repo.IAuthorRepository, readerRepo repo.IReaderRepository, interactClient interact.InteractionClient, producer producer.Producer) *ArticleLogic {
	return &ArticleLogic{authorRepo: authorRepo, readerRepo: readerRepo, interactRpc: interactClient, producer: producer}
}

func (l *ArticleLogic) Edit(ctx context.Context, req *types.EditArticleReq) (int64, error) {
	id := user.GetUidByCtx(ctx)

	artDomain := &domain.Article{
		Id:      int64(req.Id),
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: int64(id),
		},
	}
	return l.authorRepo.Save(ctx, artDomain)
}

func (l *ArticleLogic) Publish(ctx context.Context, req *types.EditArticleReq) (id int64, err error) {
	defer func() {
		// 发送文章创建事件
		if req.Id == 0 && err == nil {
			msg, _ := decodeReadEvt(ctx, id)
			err = l.producer.SendSync(ctx, domain.ArticleCreateTopic, string(msg))
		}
	}()
	uid := user.GetUidByCtx(ctx)

	artDomain := &domain.Article{
		Id:      int64(req.Id),
		Title:   req.Title,
		Content: req.Content,
		Status:  domain.ArticlePublishedStatus,
		Author: domain.Author{
			Id: uid,
		},
	}

	//保存到制作库,并刷新缓存
	artId, err := l.authorRepo.Save(ctx, artDomain)
	if err != nil {
		return 0, err
	}

	artDomain.Id = artId
	for i := 0; i < 3; i++ {
		if id, err = l.readerRepo.Save(ctx, artDomain); err == nil {
			return id, err
		}
	}

	return 0, err
}

func (l *ArticleLogic) Page(ctx context.Context, req *types.ArticlePageReq) ([]*domain.Article, error) {
	id := user.GetUidByCtx(ctx)

	return l.authorRepo.SelectPage(ctx, id, req.Page, req.Size)
}

func (l *ArticleLogic) ViewArticle(ctx context.Context, req *types.ArticleViewReq) (article *domain.Article, err error) {
	defer func() {
		if err == nil {
			jsonStr, err := decodeReadEvt(ctx, article.Id)
			if err != nil {
				logx.Errorf("反序列化读事件异常：%s", err)
				return
			}
			l.producer.SendAsync(
				ctx,
				domain.ReadTopic,
				string(jsonStr),
				func(err error) {
					logx.WithContext(ctx).Errorf("向消息队列推送文章阅读事件失败,%s", err)
				},
			)
		}
	}()
	if req.Published > 0 {
		stat, err := l.interactRpc.QueryInteractionInfo(ctx, &interact.QueryInteractionReq{Biz: domain.Biz, BizId: req.Id})
		if err != nil {
			logx.WithContext(ctx).Errorf("[RPC失败] 访问文章统计数量错误,原因:%s", err)
			return nil, errors.Wrapf(err, "用户%s访问文章计数信息失败", ctx.Value("id"))
		}
		article, err = l.readerRepo.FindById(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		article.CollectCnt = stat.CollectCnt
		article.LikeCnt = stat.LikeCnt
		article.ReadCnt = stat.ReadCnt
		return article, nil
	}
	return l.authorRepo.FindArticleById(ctx, req.Id)
}

func decodeReadEvt(ctx context.Context, bid int64) ([]byte, error) {
	id := user.GetUidByCtx(ctx)
	evt := domain.ReadEvent{
		Biz:   "article",
		BizId: bid,
		Uid:   id,
	}
	return json.Marshal(&evt)
}
