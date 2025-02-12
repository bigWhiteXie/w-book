package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"codexie.com/w-book-interact/internal/domain"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type CommentCache interface {
	// 获取根评论缓存
	GetRootComments(ctx context.Context, biz string, bizID int64, offset, size int) ([]*domain.Comment, error)
	// 设置根评论缓存
	SetRootComments(ctx context.Context, biz string, bizID int64, comments []*domain.Comment, expireDuration time.Duration) error
	// 获取子评论缓存
	GetChildComments(ctx context.Context, rootID int64, offset, size int) ([]*domain.Comment, error)
	// 设置子评论缓存
	SetChildComments(ctx context.Context, rootID int64, comments []*domain.Comment, expireDuration time.Duration) error
	// 删除根评论缓存
	DelRootComments(ctx context.Context, biz string, bizID int64) error
	// 删除子评论缓存
	DelChildComments(ctx context.Context, rootID int64) error
}

type CommentRedisCache struct {
	client *redis.Client
}

func NewCommentRedisCache(client *redis.Client) CommentCache {
	return &CommentRedisCache{client: client}
}

func (c *CommentRedisCache) GetRootComments(ctx context.Context, biz string, bizID int64, offset, size int) ([]*domain.Comment, error) {
	key := c.rootCommentKey(biz, bizID)
	return c.getCommentsFromCache(ctx, key, offset, size)
}

func (c *CommentRedisCache) SetRootComments(ctx context.Context, biz string, bizID int64, comments []*domain.Comment, expireDuration time.Duration) error {
	key := c.rootCommentKey(biz, bizID)
	return c.setCommentsToCache(ctx, key, comments, expireDuration)
}

func (c *CommentRedisCache) GetChildComments(ctx context.Context, rootID int64, offset, size int) ([]*domain.Comment, error) {
	key := c.childCommentKey(rootID)
	return c.getCommentsFromCache(ctx, key, offset, size)
}

func (c *CommentRedisCache) SetChildComments(ctx context.Context, rootID int64, comments []*domain.Comment, expireDuration time.Duration) error {
	key := c.childCommentKey(rootID)
	return c.setCommentsToCache(ctx, key, comments, expireDuration)
}

func (c *CommentRedisCache) DelRootComments(ctx context.Context, biz string, bizID int64) error {
	key := c.rootCommentKey(biz, bizID)
	return c.client.Del(ctx, key).Err()
}

func (c *CommentRedisCache) DelChildComments(ctx context.Context, rootID int64) error {
	key := c.childCommentKey(rootID)
	return c.client.Del(ctx, key).Err()
}

func (c *CommentRedisCache) rootCommentKey(biz string, bizID int64) string {
	return fmt.Sprintf("hot:comment:%s:%d", biz, bizID)
}

func (c *CommentRedisCache) childCommentKey(rootID int64) string {
	return fmt.Sprintf("hot:childComment:%d", rootID)
}

func (c *CommentRedisCache) getCommentsFromCache(ctx context.Context, key string, offset, size int) ([]*domain.Comment, error) {
	results, err := c.client.LRange(ctx, key, int64(offset), int64(offset+size-1)).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "获取评论缓存失败, key=%s", key)
	}

	var comments []*domain.Comment
	for _, result := range results {
		var comment domain.Comment
		if err := json.Unmarshal([]byte(result), &comment); err != nil {
			return nil, errors.Wrapf(err, "反序列化评论失败, key=%s", key)
		}
		comments = append(comments, &comment)
	}

	return comments, nil
}

func (c *CommentRedisCache) setCommentsToCache(ctx context.Context, key string, comments []*domain.Comment, expireDuration time.Duration) error {
	// 先删除旧缓存
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return errors.Wrapf(err, "删除旧评论缓存失败, key=%s", key)
	}

	// 序列化并存储新评论
	var args []interface{}
	for _, comment := range comments {
		data, err := json.Marshal(comment)
		if err != nil {
			return errors.Wrapf(err, "序列化评论失败, key=%s", key)
		}
		args = append(args, data)
	}

	// 使用RPush存储评论列表
	if err := c.client.RPush(ctx, key, args...).Err(); err != nil {
		return errors.Wrapf(err, "存储评论缓存失败, key=%s", key)
	}

	// 设置缓存过期时间
	if err := c.client.Expire(ctx, key, expireDuration).Err(); err != nil {
		return errors.Wrapf(err, "设置评论缓存过期时间失败, key=%s", key)
	}

	return nil
}
