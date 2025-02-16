package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-interact/internal/domain"
	"gorm.io/gorm"
)

type Comment struct {
	ID       int64         `gorm:"column:id;primary_key;auto_increment"`
	Uid      int64         `gorm:"column:uid;not null"`
	Biz      string        `gorm:"column:biz;not null;index:idx_biz_bizid_root,priority:1"`
	BizID    int64         `gorm:"column:biz_id;not null;index:idx_biz_bizid_root,priority:2"`
	RootID   int64         `gorm:"column:root_id;not null;index:idx_biz_bizid_root,priority:3;index:idx_root_ctime,priority:1"`
	ParentID sql.NullInt64 `gorm:"column:parent_id"`
	ChildNum int64         `gorm:"column:child_num;default:0"` // 子评论数量(根节点有)
	Status   int           `gorm:"column:status;default:0"`
	LikeCnt  int           `gorm:"column:like_cnt;default:0"`
	Score    int64         `gorm:"column:score;default:0"` // 根评论分数
	Content  string        `gorm:"column:content;not null"`
	Ctime    int64         `gorm:"column:created_at;not null;index:idx_root_ctime,priority:2"`
}

type DeleteCommentResult struct {
	DeletedCount int64   // 删除的评论数量
	DeletedIDs   []int64 // 被删除的评论ID列表
}

func (Comment) TableName() string {
	return "comment"
}

type CommentDAO struct {
	db *gorm.DB
}

func NewCommentDAO(db *gorm.DB) *CommentDAO {
	return &CommentDAO{db: db}
}

// GetRecentComments 获取最近的n条根评论，每个根评论最多m条子评论
func (dao *CommentDAO) GetRecentComments(ctx context.Context, biz string, bizID int64, n, m int) ([]*Comment, error) {
	// 1. 先查询根评论
	var rootComments []*Comment
	if err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND parent_id IS NULL", biz, bizID).
		Order("created_at DESC").
		Limit(n).
		Find(&rootComments).Error; err != nil {
		return nil, err
	}

	if len(rootComments) == 0 {
		return nil, nil
	}

	return rootComments, nil
}

// GetChildCommentsByRootIDs 获取子评论，按root_id分组，按created_at倒序，取前m条
func (dao *CommentDAO) GetChildCommentsByRootIDs(ctx context.Context, rootId int64, m int) ([]*Comment, error) {
	var subComments []*Comment
	if err := dao.db.WithContext(ctx).
		Where("root_id = ? and parent_id is not null", rootId).
		Order("created_at DESC").
		Limit(m).
		Find(&subComments).Error; err != nil {
		return nil, err
	}
	return subComments, nil
}

func (dao *CommentDAO) GetCommentByID(ctx context.Context, id int64) (*Comment, error) {
	var comment Comment
	if err := dao.db.WithContext(ctx).
		Where("id = ? AND status = 0", id). // 排除已删除的评论
		First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: comment id %d", gorm.ErrRecordNotFound, id)
		}
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}
	return &comment, nil
}

// DeleteComment 删除评论
// 若为根评论则删除所有子评论，否则仅将自己的status设置为0
func (dao *CommentDAO) DeleteComment(ctx context.Context, id int64, uid int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 获取评论信息
		var target Comment
		if err := tx.First(&target, id).Error; err != nil {
			return err
		}

		if target.Uid != uid {
			return codeerr.LogCodeError(ctx, "无权删除该评论", "uid=%d无权删除该评论 comment_id=%d", uid, id)
		}

		// 2. 判断是否为根评论
		if target.RootID == target.ID { // 根评论
			// 删除所有root_id指向它的评论
			if err := tx.Where("root_id = ?", target.ID).
				Delete(&Comment{}).Error; err != nil {
				return err
			}
		} else { // 子评论
			// 仅将当前评论的status设置为0
			if err := tx.Model(&target).
				UpdateColumn("status", 2).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// CreateComment 创建评论,若根评论不存在则插入失败
func (dao *CommentDAO) CreateComment(ctx context.Context, comment *Comment) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 如果是子评论，验证父评论存在性并设置RootID
		if comment.ParentID.Valid {
			var parent Comment
			// 查询父节点，同时检查status=1
			if err := tx.Where("id = ? AND status = 1", comment.ParentID.Int64).First(&parent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("父评论不存在或已被删除: %d", comment.ParentID.Int64)
				}
				return err
			}

			// 设置RootID为父评论的RootID
			comment.RootID = parent.RootID

			// 更新根评论的child_num，并检查rowsAffected
			result := tx.Model(&Comment{}).
				Where("id = ?", comment.RootID).
				UpdateColumn("child_num", gorm.Expr("child_num + 1"))
			if result.Error != nil {
				return fmt.Errorf("更新子评论数量失败: %w", result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("根评论不存在或已被删除: %d", comment.RootID)
			}
		}

		// 2. 创建评论
		if err := tx.Create(comment).Error; err != nil {
			return fmt.Errorf("创建评论失败: %w", err)
		}

		// 3. 如果是根评论，设置RootID为自己的ID
		if !comment.ParentID.Valid {
			comment.RootID = comment.ID
			if err := tx.Model(comment).Update("root_id", comment.ID).Error; err != nil {
				return fmt.Errorf("更新根评论ID失败: %w", err)
			}
		}

		return nil
	})
}

func (dao *CommentDAO) CreateCommentV2(ctx context.Context, comment *Comment) error {
	if err := dao.db.WithContext(ctx).Create(comment).Error; err != nil {
		return fmt.Errorf("创建评论失败: %w", err)
	}
	if comment.RootID == 0 {
		comment.RootID = comment.ID
		if err := dao.db.WithContext(ctx).Model(comment).Update("root_id", comment.ID).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetRootCommentsByScore 分页获取根评论（按score倒序）
// lastScore 上次查询的最小score，首次查询传0表示从最大score开始
func (dao *CommentDAO) GetRootCommentsByScore(ctx context.Context, biz string, bizID int64, lastScore int64, pageSize int) ([]*Comment, error) {
	var comments []*Comment
	query := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND parent_id IS NULL", biz, bizID).
		Order("score DESC,id DESC").
		Limit(pageSize)

	if lastScore > 0 {
		query = query.Where("score < ?", lastScore)
	}

	if err := query.Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("分页查询根评论失败: %w", err)
	}
	return comments, nil
}

// GetChildCommentsByTime 分页获取子评论（按时间倒序）
// lastTime 上次查询的最早时间戳，首次查询传0表示从最新开始
func (dao *CommentDAO) GetChildCommentsByTime(ctx context.Context, rootId int64, lastTime int64, pageSize int) ([]*Comment, error) {
	var comments []*Comment
	query := dao.db.WithContext(ctx).
		Where("root_id = ? AND id != ?", rootId, rootId). // 排除根评论自身
		Order("created_at DESC").                         // 按创建时间倒序
		Limit(pageSize)

	if lastTime > 0 {
		query = query.Where("created_at < ?", lastTime)
	}

	if err := query.Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (dao *CommentDAO) LikeComment(ctx context.Context, id int64, isLike bool) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 查询评论是否存在
		var comment Comment
		if err := tx.First(&comment, id).Error; err != nil {
			return err
		}

		// 2. 更新点赞数
		var updateValue int
		if isLike {
			updateValue = 1
		} else {
			updateValue = -1
		}

		// 3. 执行更新
		if err := tx.Model(&comment).
			UpdateColumn("like_cnt", gorm.Expr("like_cnt + ?", updateValue)).Error; err != nil {
			return err
		}

		return nil
	})
}

func CommentDo2Domain(ctx context.Context, comment *Comment) *domain.Comment {
	return &domain.Comment{
		ID:       comment.ID,
		Uid:      comment.Uid,
		Biz:      comment.Biz,
		BizID:    comment.BizID,
		ChildNum: comment.ChildNum,
		LikeCnt:  comment.LikeCnt,
		ParentID: comment.ParentID.Int64,
		Content:  comment.Content,
		Childs:   []*domain.Comment{},
		Ctime:    time.UnixMilli(comment.Ctime),
		Score:    comment.Score,
		RootID:   comment.RootID,
	}
}

func CommentDomain2Do(ctx context.Context, comment *domain.Comment) *Comment {
	return &Comment{
		ID:       comment.ID,
		Uid:      comment.Uid,
		Biz:      comment.Biz,
		ParentID: sql.NullInt64{Int64: comment.ParentID, Valid: comment.ParentID != 0},
		BizID:    comment.BizID,
		ChildNum: comment.ChildNum,
		LikeCnt:  comment.LikeCnt,
		Content:  comment.Content,
		Score:    comment.Score,
		RootID:   comment.RootID,
	}
}

// 新增方法：分页获取多个资源的根评论
func (dao *CommentDAO) GetRootCommentsByBizIDs(ctx context.Context, biz string, bizIDs []int64, size int) ([]*Comment, error) {
	var comments []*Comment

	query := dao.db.WithContext(ctx).Raw(`
		SELECT * FROM (
			SELECT *,
				ROW_NUMBER() OVER (
					PARTITION BY biz_id 
					ORDER BY score DESC
				) AS rn
			FROM comment
			WHERE biz = ? 
				AND biz_id IN (?)
				AND parent_id IS NULL
		) t WHERE rn <= ?
	`, biz, bizIDs, size)

	if err := query.Scan(&comments).Error; err != nil {
		return nil, fmt.Errorf("分资源分页查询失败: %w", err)
	}

	return comments, nil
}

func (r *CommentDAO) RefreshCommentScore(ctx context.Context, rootID int64) error {
	return r.db.WithContext(ctx).
		Model(&Comment{}).
		Where("id = ?", rootID).
		Update("score", gorm.Expr("like_cnt * 2 + child_num")).Error
}
