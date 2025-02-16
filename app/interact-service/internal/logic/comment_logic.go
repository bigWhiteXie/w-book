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
	userId, ok := ctx.Value("id").(int)
	if !ok {
		return nil, codeerr.LogCodeError(ctx, "认证异常", "用户id不存在")
	}
	uid := int64(userId)

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
		return nil, codeerr.LogCodeError(ctx, "发送评论事件失败", "EvtMsg=%s", msgJson)
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
	// 筛选出没有点赞的评论，剩下的评论点赞状态走缓存+数据库
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

	if err := l.commentRepo.DeleteComment(ctx, commentID, uid); err != nil {
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
	l.commentLikeFilter.AddNoCreate(ctx, fmt.Sprintf("comment:like_user:%d", commentID), time.Hour*24, uid)
	go func() {
		msg := domain.CommentLikeEvent{
			CommentID: commentID,
			IsLike:    isLike,
			Uid:       uid,
		}

		msgJson, _ := json.Marshal(msg)
		if err := l.kafkaProducer.SendSync(ctx, domain.CommentLikeEvtTopic, string(msgJson)); err != nil {
			codeerr.LogCodeError(ctx, "发送评论事件失败", "EvtMsg=%s", msgJson)
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
