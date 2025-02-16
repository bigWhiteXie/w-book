package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-common/kafka/producer"
	"codexie.com/w-book-common/middleware/filter"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/types"
	"github.com/redis/go-redis/v9"
)

type CommentLogic struct {
	commentRepo       repo.ICommentRepo
	kafkaProducer     producer.Producer
	commentLikeFilter *filter.RedisBloomFilter
}

func NewCommentLogic(commentRepo repo.ICommentRepo, kafkaProducer producer.Producer, redisClient *redis.Client) *CommentLogic {
	commentLikeFilter := filter.NewRedisBloomFilter(redisClient, 10000000, 10)
	return &CommentLogic{commentRepo: commentRepo, kafkaProducer: kafkaProducer, commentLikeFilter: commentLikeFilter}
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
	keys := make([]string, 0, len(comments))
	for _, c := range comments {
		keys = append(keys, fmt.Sprintf("comment:like_user:%d", c.ID))
	}

	filterStatus, _ := l.commentLikeFilter.JudgeKeys(ctx, keys, uid)
	statusMap := make(map[int64]bool, len(comments))
	cids := make([]int64, 0, len(comments))
	for i, status := range filterStatus {
		if status == filter.KeyNotExist || status == filter.ValExist {
			cids = append(cids, comments[i].ID)
		} else {
			statusMap[comments[i].ID] = false
		}
	}

	likeStatus, err := l.commentRepo.GetLikeStatusByUid(ctx, uid, cids)
	if err != nil {
		return nil, err
	}
	for _, c := range comments {
		c.IsLike = likeStatus[c.ID] || statusMap[c.ID]
		for _, cc := range c.Childs {
			cc.IsLike = likeStatus[cc.ID] || statusMap[cc.ID]
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
	uid := int64(ctx.Value("id").(int))
	cids := make([]int64, 0, len(comments))
	for _, c := range comments {
		cids = append(cids, c.ID)
	}
	likeStatus, err := l.commentRepo.GetLikeStatusByUid(ctx, uid, cids)
	if err != nil {
		return nil, err
	}
	for _, c := range comments {
		c.IsLike = likeStatus[c.ID]
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

// 查询点赞过该评论的用户id
func (l *CommentLogic) GetCommentLikeUserIDs(ctx context.Context, commentID int64) ([]int64, error) {
	return l.commentRepo.GetCommentLikeUserIDs(ctx, commentID)
}

func (l *CommentLogic) LikeComment(ctx context.Context, commentID int64, isLike bool) error {
	uid := int64(ctx.Value("id").(int))
	err := l.commentRepo.LikeComment(ctx, commentID, uid, isLike)
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

// 将该评论添加点赞过的用户id添加到redis中
func (l *CommentLogic) AddCommentFilter(ctx context.Context, commentID int64, uids []int64, expire time.Duration) error {
	key := fmt.Sprintf("comment:like_user:%d", commentID)
	return l.commentLikeFilter.Add(ctx, key, expire, uids)
}
