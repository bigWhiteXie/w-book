package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Comment struct {
	ID       int64         `gorm:"column:id;primary_key;auto_increment"`
	Biz      string        `gorm:"column:biz;not null;index:idx_biz_bizid_root,priority:1"`
	BizID    int64         `gorm:"column:biz_id;not null;index:idx_biz_bizid_root,priority:2"`
	RootID   int64         `gorm:"column:root_id;not null;index:idx_biz_bizid_root,priority:3;index:idx_root_ctime,priority:1"`
	ParentID sql.NullInt64 `gorm:"column:parent_id"`
	ChildNum int64         `gorm:"column:child_num;default:0"` // 子评论数量(根节点有)
	Status   int           `gorm:"column:status;default:1"`    // 是否可见
	Score    int64         `gorm:"column:score;default:0"`     // 根评论分数
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

func (dao *CommentDAO) GetChildCommentsByRootIDs(ctx context.Context, rootIDs []int64, m int) ([]*Comment, error) {
	var subComments []*Comment
	if err := dao.db.WithContext(ctx).Raw(`
		SELECT * FROM (
			SELECT *,
				ROW_NUMBER() OVER (
					PARTITION BY root_id 
					ORDER BY created_at DESC
				) AS rn
			FROM comment
			WHERE root_id IN ?
				AND parent_id IS NOT NULL
		) t WHERE rn <= ?
	`, rootIDs, m).Scan(&subComments).Error; err != nil {
		return nil, err
	}
	return subComments, nil
}

func (dao *CommentDAO) GetCommentByID(ctx context.Context, id int64) (*Comment, error) {
	var comment Comment
	if err := dao.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

// DeleteComment 删除评论
// 若为根评论则删除所有子评论，否则仅将自己的status设置为0
func (dao *CommentDAO) DeleteComment(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 获取评论信息
		var target Comment
		if err := tx.First(&target, id).Error; err != nil {
			return err
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
				UpdateColumn("status", 0).Error; err != nil {
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

	return nil
}
