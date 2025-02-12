package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCommentDAO_CreateCommentV2(t *testing.T) {
	db := setupTestDB(t)
	dao := NewCommentDAO(db)

	t.Run("创建根评论", func(t *testing.T) {
		comment := &Comment{
			Biz:     "test",
			BizID:   1,
			Content: "根评论",
		}

		err := dao.CreateComment(context.Background(), comment)
		assert.NoError(t, err)
		assert.Equal(t, comment.ID, comment.RootID)
	})

	t.Run("创建子评论", func(t *testing.T) {
		// 先创建根评论
		root := &Comment{
			Biz:     "test",
			BizID:   1,
			Content: "根评论",
		}
		err := dao.CreateComment(context.Background(), root)
		assert.NoError(t, err)

		// 创建子评论
		child := &Comment{
			Biz:      "test",
			BizID:    1,
			ParentID: sql.NullInt64{Int64: root.ID, Valid: true},
			Content:  "子评论",
		}
		err = dao.CreateComment(context.Background(), child)
		assert.NoError(t, err)
		assert.Equal(t, root.ID, child.RootID)
	})

	t.Run("父评论不存在", func(t *testing.T) {
		comment := &Comment{
			Biz:      "test",
			BizID:    1,
			ParentID: sql.NullInt64{Int64: 999, Valid: true},
			Content:  "子评论",
		}

		err := dao.CreateComment(context.Background(), comment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "父评论不存在或已被删除")
	})

	t.Run("根评论不存在", func(t *testing.T) {
		root := &Comment{
			Biz:      "test",
			BizID:    1,
			Content:  "根评论",
			ParentID: sql.NullInt64{Int64: 0, Valid: false},
		}
		err := dao.CreateComment(context.Background(), root)
		assert.NoError(t, err)

		// 创建父评论
		parent := &Comment{
			Biz:      "test",
			BizID:    1,
			Content:  "父评论",
			ParentID: sql.NullInt64{Int64: root.ID, Valid: true},
		}
		err = dao.CreateComment(context.Background(), parent)
		assert.NoError(t, err)

		// 删除根评论
		err = dao.DeleteComment(context.Background(), root.ID)
		assert.NoError(t, err)

		// 尝试创建子评论
		child := &Comment{
			Biz:      "test",
			BizID:    1,
			ParentID: sql.NullInt64{Int64: root.ID, Valid: true},
			Content:  "子评论",
			RootID:   parent.ID,
		}
		err = dao.CreateComment(context.Background(), child)
		assert.Error(t, err)
		// assert.Contains(t, err.Error(), "根评论不存在或已被删除")
	})
}

func TestCommentDAO_DeleteCommentV2(t *testing.T) {
	db := setupTestDB(t)
	dao := NewCommentDAO(db)

	t.Run("删除根评论", func(t *testing.T) {
		// 创建根评论
		root := &Comment{
			Biz:     "test",
			BizID:   1,
			Content: "根评论",
		}
		err := dao.CreateComment(context.Background(), root)
		assert.NoError(t, err)

		// 创建子评论
		child := &Comment{
			Biz:      "test",
			BizID:    1,
			ParentID: sql.NullInt64{Int64: root.ID, Valid: true},
			Content:  "子评论",
		}
		err = dao.CreateComment(context.Background(), child)
		assert.NoError(t, err)

		// 删除根评论
		err = dao.DeleteComment(context.Background(), root.ID)
		assert.NoError(t, err)

		// 验证根评论和子评论都被删除
		var rootComment Comment
		err = db.First(&rootComment, root.ID).Error
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

		var childComment Comment
		err = db.First(&childComment, child.ID).Error
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("删除子评论", func(t *testing.T) {
		// 创建根评论
		root := &Comment{
			Biz:     "test",
			BizID:   1,
			Content: "根评论",
		}
		err := dao.CreateComment(context.Background(), root)
		assert.NoError(t, err)

		// 创建子评论
		child := &Comment{
			Biz:      "test",
			BizID:    1,
			ParentID: sql.NullInt64{Int64: root.ID, Valid: true},
			Content:  "子评论",
		}
		err = dao.CreateComment(context.Background(), child)
		assert.NoError(t, err)

		// 删除子评论
		err = dao.DeleteComment(context.Background(), child.ID)
		assert.NoError(t, err)

		// 验证子评论被删除，根评论仍然存在
		var deletedChild Comment
		err = db.First(&deletedChild, child.ID).Error
		assert.NoError(t, err)
		assert.True(t, deletedChild.Status == 0)

		var rootComment Comment
		err = db.First(&rootComment, root.ID).Error
		assert.NoError(t, err)
	})
}
