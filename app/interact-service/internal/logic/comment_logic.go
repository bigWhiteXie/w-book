package logic

import (
	"context"
	"encoding/json"

	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-common/kafka/producer"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/types"
)

type CommentLogic struct {
	commentRepo   repo.ICommentRepo
	kafkaProducer producer.Producer
}

func NewCommentLogic(commentRepo repo.ICommentRepo, kafkaProducer producer.Producer) *CommentLogic {
	return &CommentLogic{commentRepo: commentRepo, kafkaProducer: kafkaProducer}
}

// AddComment 添加评论
func (l *CommentLogic) AddComment(ctx context.Context, req *types.AddCommentReq) (*domain.Comment, error) {
	uid := int64(ctx.Value("id").(int))

	comment, err := l.commentRepo.CreateComment(ctx, req.Biz, req.BizID, req.ParentID, req.Content, uid)
	if err != nil {
		return nil, err
	}
	msg := domain.CommentEvent{
		CommentID: comment.ID,
		RootID:    comment.RootID,
		Action:    domain.CreateCommentEvt,
	}
	msgJson, _ := json.Marshal(msg)
	if err := l.kafkaProducer.SendSync(ctx, domain.CommentEvtTopic, string(msgJson)); err != nil {
		return nil, codeerr.LogCodeError(ctx, "发送评论事件失败", err, "EvtMsg=%s", msgJson)
	}
	return comment, nil
}

// GetRootComments 获取根评论列表(每条根评论附带子评论)
func (l *CommentLogic) GetRootComments(ctx context.Context, req *types.GetRootCommentsReq) ([]*domain.Comment, error) {
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}
	uid := int64(ctx.Value("id").(int))
	comments, err := l.commentRepo.GetRootComments(ctx, req.Biz, req.BizID, req.Offset, req.Size, req.LastScore)
	if err != nil {
		return nil, err
	}
	cids := make([]int64, 0, len(comments))
	for _, c := range comments {
		cids = append(cids, c.ID)
	}
	likeStatus, err := l.commentRepo.GetLikeStatus(ctx, uid, cids)
	if err != nil {
		return nil, err
	}
	for _, c := range comments {
		c.IsLike = likeStatus[c.ID]
		for _, cc := range c.Childs {
			cc.IsLike = likeStatus[cc.ID]
		}
	}

	return comments, nil
}

// GetChildComments 获取子评论列表
func (l *CommentLogic) GetChildComments(ctx context.Context, req *types.GetChildCommentsReq) ([]*domain.Comment, error) {
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	comments, err := l.commentRepo.GetChildComments(ctx, req.RootID, req.LastCTime, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// DeleteComment 删除评论
func (l *CommentLogic) DeleteComment(ctx context.Context, commentID int64) error {
	uid := int64(ctx.Value("id").(int))

	// 先查询评论是否存在且属于当前用户
	comment, err := l.commentRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}

	if comment.Uid != uid {
		return codeerr.LogCodeError(ctx, "无权删除该评论", err, "uid=%d无权删除该评论 comment_id=%d", uid, commentID)
	}

	if err := l.commentRepo.DeleteComment(ctx, commentID); err != nil {
		return err
	}
	return nil
}

func (l *CommentLogic) LikeComment(ctx context.Context, commentID int64) error {
	uid := int64(ctx.Value("id").(int))
	err := l.commentRepo.LikeComment(ctx, commentID, uid)
	if err != nil {
		return err
	}
	go func() {
		msg := domain.CommentEvent{
			CommentID: commentID,
			Action:    domain.LikeCommentEvt,
		}
		comment, err := l.commentRepo.GetCommentByID(ctx, commentID)
		if err != nil {
			codeerr.LogCodeError(ctx, "获取评论失败导致未能发送点赞事件，需重新执行流程", err, "commentID=%d", commentID)
			// todo: 告警
			return
		}
		if comment.RootID != commentID {
			return
		}
		msg.RootID = comment.RootID
		msgJson, _ := json.Marshal(msg)
		if err := l.kafkaProducer.SendSync(ctx, domain.CommentEvtTopic, string(msgJson)); err != nil {
			codeerr.LogCodeError(ctx, "发送评论事件失败", err, "EvtMsg=%s", msgJson)
		}
	}()
	return nil
}
func (l *CommentLogic) GetRootCommentsByBizIDs(ctx context.Context, biz string, bizIDs []int64, size int) ([]*domain.Comment, error) {
	return l.commentRepo.GetRootCommentsByBizIDs(ctx, biz, bizIDs, size)
}

// GetCommentByID 获取单个评论详情
func (l *CommentLogic) GetCommentByID(ctx context.Context, commentID int64) (*domain.Comment, error) {
	comment, err := l.commentRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	return comment, nil
}
