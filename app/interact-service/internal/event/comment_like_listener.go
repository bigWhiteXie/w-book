package event

import (
	"context"

	"codexie.com/w-book-common/kafka/consumer"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

type CommentLikeEvtListener struct {
	commentRepo   repo.ICommentRepo
	batchConsumer *consumer.BatchConsumer[domain.CommentLikeEvent]
}

func NewCommentLikeEvtListener(client sarama.Client, commentRepo repo.ICommentRepo) *CommentLikeEvtListener {
	l := &CommentLikeEvtListener{
		commentRepo: commentRepo,
	}
	l.batchConsumer = consumer.NewBatchConsumer[domain.CommentLikeEvent](
		domain.CommentLikeEvtTopic,
		client,
		"interact-comment-group",
		50,
		l.handleBatchComment,
	)
	return l
}

func (c *CommentLikeEvtListener) handleBatchComment(events []domain.CommentLikeEvent, msgs []*sarama.ConsumerMessage) error {
	ctx := context.Background()
	logx.WithContext(ctx).Infof("收到批量评论事件，数量:%d", len(events))

	// 处理每个评论事件
	for _, evt := range events {
		c.commentRepo.HandleLikeCommentEvent(ctx, evt.CommentID, evt.IsLike, evt.Uid)
	}
	return nil
}

func (c *CommentLikeEvtListener) Start() {
	c.batchConsumer.StartListner()
}

func (c *CommentLikeEvtListener) Stop() {
	c.batchConsumer.Stop()
}
