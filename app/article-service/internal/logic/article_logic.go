package logic

import (
	"context"
	"encoding/json"

	"codexie.com/w-book-article/internal/domain"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-common/producer"
	"codexie.com/w-book-common/user"
	"codexie.com/w-book-interact/api/pb/interact"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	articleBiz = "article"
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

func (l *ArticleLogic) Edit(ctx context.Context, artDomain *domain.Article) (int64, error) {
	return l.authorRepo.Save(ctx, artDomain)
}

func (l *ArticleLogic) Publish(ctx context.Context, artDomain *domain.Article) (id int64, err error) {
	var published = true
	log := logx.WithContext(ctx)
	defer func() {
		// 发送文章创建事件
		if !published && err == nil {
			log.Infof("ArticleLogic] 创建文章发布事件, id:%d", artDomain.Id)
			msg, _ := decodeReadEvt(ctx, id)
			err = l.producer.SendSync(ctx, domain.ArticleCreateTopic, string(msg))
		}
	}()

	if artDomain.Id != 0 {
		if art, err := l.readerRepo.FindById(ctx, artDomain.Id, false); err != nil {
			log.Errorf("Fail to find published article by id %d,cause:%s", artDomain.Id, err)
		} else if art == nil || art.Id == 0 {
			published = false
		}
	}
	//保存到制作库,并刷新缓存
	artId, err := l.authorRepo.Save(ctx, artDomain)
	if err != nil {
		return 0, errors.WithMessage(err, "[ArticleLogic] 保存文章到制作库失败")
	}

	artDomain.Id = artId
	for i := 0; i < 3; i++ {
		if id, err = l.readerRepo.Save(ctx, artDomain); err == nil {
			return id, nil
		}
	}

	return 0, err
}

func (l *ArticleLogic) Page(ctx context.Context, uid int64, page, size int) ([]*domain.Article, error) {
	return l.authorRepo.SelectPage(ctx, uid, page, size)
}

func (l *ArticleLogic) ViewArticle(ctx context.Context, id int64, published bool) (article *domain.Article, err error) {
	log := logx.WithContext(ctx)
	defer func() {
		if err == nil {
			jsonStr, err := decodeReadEvt(ctx, article.Id)
			if err != nil {
				log.Errorf("[ArticleLogic_ViewArticle] 反序列化读事件异常：%s", err)
				return
			}
			l.producer.SendAsync(
				ctx,
				domain.ReadTopic,
				string(jsonStr),
				func(err error) {
					log.Errorf("向消息队列推送文章阅读事件失败,%s", err)
				},
			)
		}
	}()
	if published {
		uid := user.GetUidByCtx(ctx)
		stat, err := l.interactRpc.QueryInteractionInfo(ctx, &interact.QueryInteractionReq{Uid: uid, Biz: domain.Biz, BizId: id})
		if err != nil {
			return nil, errors.Wrapf(err, "[ArticleLogic_ViewArticle]Rpc访问交互信息异常,uid=%d,biz=%s,bizId=%d", ctx.Value("id"), domain.Biz, id)
		}
		article, err = l.readerRepo.FindById(ctx, id, true)
		if err != nil {
			return nil, err
		}
		article.CollectCnt = stat.CollectCnt
		article.LikeCnt = stat.LikeCnt
		article.ReadCnt = stat.ReadCnt
		article.IsCollected = stat.IsCollected
		article.IsLiked = stat.IsLiked

		return article, nil
	}
	return l.authorRepo.FindArticleById(ctx, id)
}

func (l *ArticleLogic) GetTopLikeArticles(ctx context.Context) ([]*domain.Article, error) {
	topResp, err := l.interactRpc.TopLike(ctx, &interact.TopLikeReq{Biz: articleBiz})
	if err != nil {
		return nil, errors.Wrapf(err, "[ArticleLogic_GetTopLikeArticles]Rpc调用interactRpc的TopLike方法异常,biz=%s", articleBiz)
	}
	return l.readerRepo.GetShortArticles(ctx, topResp.Items)
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
