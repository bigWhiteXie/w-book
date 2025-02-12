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

	// 2. 获取所有根评论的ID
	rootIDs := make([]int64, len(rootComments))
	for i, c := range rootComments {
		rootIDs[i] = c.ID
	}

	// 3. 查询子评论
	// 3. 使用窗口函数查询子评论
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

	result := append(rootComments, subComments...)
	return result, nil
}

// DeleteComment 删除评论
func (dao *CommentDAO) DeleteComment(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 先获取目标评论的root_id
		var targetRootID int64
		if err := tx.Model(&Comment{}).
			Select("root_id").
			Where("id = ?", id).
			Scan(&targetRootID).Error; err != nil {
			return fmt.Errorf("查询root_id失败: %w", err)
		}
		if targetRootID == 0 {
			return nil
		}

		// 2. 锁定整个子树范围
		if err := tx.Exec(`
            SELECT * FROM comment 
            WHERE root_id = ?
            FOR UPDATE
        `, targetRootID).Error; err != nil {
			return fmt.Errorf("锁定失败: %w", err)
		}

		// 3. 执行递归查询获取所有子节点
		var commentIDs []int64
		if err := tx.Raw(`
            WITH RECURSIVE comment_tree AS (
                SELECT id FROM comment WHERE id = ?
                UNION ALL
                SELECT c.id 
                FROM comment c
                INNER JOIN comment_tree ct ON c.parent_id = ct.id
            )
            SELECT id FROM comment_tree;
        `, id).Scan(&commentIDs).Error; err != nil {
			return fmt.Errorf("查询评论树失败: %w", err)
		}

		// 4. 执行删除
		if err := tx.Where("id IN ?", commentIDs).Delete(&Comment{}).Error; err != nil {
			return fmt.Errorf("删除失败: %w", err)
		}

		return nil
	})
}

// CreateComment 创建评论
func (dao *CommentDAO) CreateComment(ctx context.Context, comment *Comment) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 如果是子评论，验证父评论存在性并设置RootID
		if comment.ParentID.Valid {
			var parent Comment
			if err := tx.First(&parent, comment.ParentID.Int64).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("父评论不存在: %d", comment.ParentID.Int64)
				}
				return err
			}
			// 设置RootID为父评论的RootID
			comment.RootID = parent.RootID
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
