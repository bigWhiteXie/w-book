package event

import (
	"context"

	"codexie.com/w-book-common/kafka/consumer"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

type CommentEvtListener struct {
	commentRepo   repo.ICommentRepo
	batchConsumer *consumer.BatchConsumer[domain.CommentEvent]
}

func NewCommentEvtListener(client sarama.Client, commentRepo repo.ICommentRepo) *CommentEvtListener {
	l := &CommentEvtListener{
		commentRepo: commentRepo,
	}
	l.batchConsumer = consumer.NewBatchConsumer[domain.CommentEvent](
		domain.CommentEvtTopic,
		client,
		"interact-comment-group",
		50,
		l.handleBatchComment,
	)
	return l
}

func (c *CommentEvtListener) handleBatchComment(events []domain.CommentEvent, msgs []*sarama.ConsumerMessage) error {
	ctx := context.Background()
	logx.WithContext(ctx).Infof("收到批量评论事件，数量:%d", len(events))

	// 处理每个评论事件
	for _, evt := range events {
		// todo: 处理幂等性

		switch evt.Action {
		case domain.CreateCommentEvt:
			c.commentRepo.HandleCommentCreateEvent(ctx, evt.CommentID, evt.RootID)

		case domain.LikeCommentEvt:
			c.commentRepo.HandleLikeCommentEvent(ctx, evt.CommentID)

		default:
			logx.WithContext(ctx).Errorf("不支持的事件类型 action:%s", evt.Action)
			continue
		}
	}
	return nil
}

func (c *CommentEvtListener) Start() {
	c.batchConsumer.StartListner()
}

func (c *CommentEvtListener) Stop() {
	c.batchConsumer.Stop()
}
