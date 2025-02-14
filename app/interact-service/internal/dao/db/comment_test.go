package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupTestDB 初始化测试数据库连接
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "root:j3391111@tcp(127.0.0.1:3306)/w-book-interaction?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// 清理表
	err = db.Migrator().DropTable(&Comment{})
	require.NoError(t, err)

	// 创建表
	err = db.AutoMigrate(&Comment{})
	require.NoError(t, err)

	return db
}

func TestCommentDAO_CreateComment(t *testing.T) {
	db := setupTestDB(t)
	dao := NewCommentDAO(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		comment *Comment
		wantErr bool
	}{
		{
			name: "create root comment",
			comment: &Comment{
				Biz:     "article",
				BizID:   1,
				Content: "root comment",
			},
			wantErr: false,
		},
		{
			name: "create child comment",
			comment: &Comment{
				Biz:     "article",
				BizID:   1,
				Content: "child comment",
				ParentID: sql.NullInt64{
					Int64: 1, // 将被设置为第一个评论的ID
					Valid: true,
				},
			},
			wantErr: false,
		},
		{
			name: "create comment with invalid parent",
			comment: &Comment{
				Biz:     "article",
				BizID:   1,
				Content: "invalid parent",
				ParentID: sql.NullInt64{
					Int64: 999, // 不存在的父评论ID
					Valid: true,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dao.CreateComment(ctx, tt.comment)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotZero(t, tt.comment.ID)

			// 验证评论是否正确保存
			var saved Comment
			err = db.First(&saved, tt.comment.ID).Error
			assert.NoError(t, err)
			assert.Equal(t, tt.comment.Content, saved.Content)

			// 验证RootID设置
			if !tt.comment.ParentID.Valid {
				assert.Equal(t, tt.comment.ID, saved.RootID)
			} else {
				var parent Comment
				err = db.First(&parent, tt.comment.ParentID.Int64).Error
				assert.NoError(t, err)
				assert.Equal(t, parent.RootID, saved.RootID)
			}
		})
	}
}

func TestCommentDAO_GetRecentComments(t *testing.T) {
	db := setupTestDB(t)
	dao := NewCommentDAO(db)
	ctx := context.Background()

	// 准备测试数据
	rootComment := &Comment{
		Biz:     "article",
		BizID:   1,
		Content: "root comment",
		Ctime:   time.Now().Unix(),
	}
	err := dao.CreateComment(ctx, rootComment)
	require.NoError(t, err)

	// 创建多个子评论
	for i := 0; i < 5; i++ {
		childComment := &Comment{
			Biz:     "article",
			BizID:   1,
			Content: fmt.Sprintf("child comment %d", i),
			ParentID: sql.NullInt64{
				Int64: rootComment.ID,
				Valid: true,
			},
			Ctime: time.Now().Add(time.Duration(i) * time.Second).Unix(),
		}
		err := dao.CreateComment(ctx, childComment)
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		biz       string
		bizID     int64
		n         int // 根评论数量
		m         int // 每个根评论的子评论数量
		wantRoot  int // 期望的根评论数量
		wantChild int // 期望的子评论数量
		wantErr   bool
	}{
		{
			name:      "get recent comments",
			biz:       "article",
			bizID:     1,
			n:         1,
			m:         3,
			wantRoot:  1,
			wantChild: 3,
			wantErr:   false,
		},
		{
			name:      "no comments found",
			biz:       "nonexistent",
			bizID:     1,
			n:         1,
			m:         3,
			wantRoot:  0,
			wantChild: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments, err := dao.GetRecentComments(ctx, tt.biz, tt.bizID, tt.n, tt.m)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 计算根评论和子评论的数量
			rootCount := 0
			childCount := 0
			for _, c := range comments {
				if !c.ParentID.Valid {
					rootCount++
				} else {
					childCount++
				}
			}

			assert.Equal(t, tt.wantRoot, rootCount)
			assert.Equal(t, tt.wantChild, childCount)

			// 验证评论排序
			lastTime := int64(1<<63 - 1) // max int64
			for _, c := range comments {
				if !c.ParentID.Valid {
					continue
				}
				assert.LessOrEqual(t, c.Ctime, lastTime)
				lastTime = c.Ctime
			}
		})
	}
}

func TestCommentDAO_DeleteComment(t *testing.T) {
	db := setupTestDB(t)
	dao := NewCommentDAO(db)
	ctx := context.Background()

	// 创建测试数据：一个根评论和多个子评论
	root := &Comment{
		Biz:     "article",
		BizID:   1,
		Content: "root comment",
	}
	err := dao.CreateComment(ctx, root)
	require.NoError(t, err)

	// // 创建两层子评论
	child1 := &Comment{
		Biz:     "article",
		BizID:   1,
		Content: "child comment 1",
		ParentID: sql.NullInt64{
			Int64: root.ID,
			Valid: true,
		},
		RootID: 1,
	}
	err = dao.CreateComment(ctx, child1)
	require.NoError(t, err)

	child2 := &Comment{
		Biz:     "article",
		BizID:   1,
		Content: "child comment 2",
		ParentID: sql.NullInt64{
			Int64: child1.ID,
			Valid: true,
		},
		RootID: 1,
	}
	err = dao.CreateComment(ctx, child2)
	require.NoError(t, err)

	tests := []struct {
		name      string
		commentID int64
		wantErr   bool
	}{
		{
			name:      "delete root comment",
			commentID: 1,
			wantErr:   false,
		},
		{
			name:      "delete non-existent comment",
			commentID: 999,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dao.DeleteComment(ctx, tt.commentID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 验证评论及其子评论是否被删除
			var count int64
			err = db.Model(&Comment{}).Where("id = ? OR root_id = ?", tt.commentID, tt.commentID).Count(&count).Error
			assert.NoError(t, err)
			assert.Zero(t, count)

		})
	}
}

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
