package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLikeInfoDao_BatchFindLikeInfo(t *testing.T) {
	// 使用comment_test.go的数据库初始化方法
	db := setupTestDB(t)
	dao := NewLikeInfoDao(db)
	ctx := context.Background()

	// 准备测试数据
	testData := []*LikeInfo{
		{Biz: "article", BizId: 1, Uid: 1001, Status: 1, Ctime: time.Now().UnixMilli()},
		{Biz: "article", BizId: 2, Uid: 1001, Status: 0, Ctime: time.Now().UnixMilli()},
		{Biz: "article", BizId: 3, Uid: 1001, Status: 1, Ctime: time.Now().UnixMilli()},
		{Biz: "comment", BizId: 1, Uid: 1001, Status: 1, Ctime: time.Now().UnixMilli()}, // 不同业务类型
	}
	require.NoError(t, db.Create(testData).Error)

	tests := []struct {
		name    string
		uid     int64
		biz     string
		bizIds  []int64
		want    map[int64]bool
		wantErr bool
	}{
		{
			name:    "正常批量查询",
			uid:     1001,
			biz:     "article",
			bizIds:  []int64{1, 2, 3, 4},
			want:    map[int64]bool{1: true, 2: false, 3: true, 4: false},
			wantErr: false,
		},
		{
			name:    "空bizIds列表",
			uid:     1001,
			biz:     "article",
			bizIds:  []int64{},
			want:    map[int64]bool{},
			wantErr: false,
		},
		{
			name:    "不同业务类型",
			uid:     1001,
			biz:     "comment",
			bizIds:  []int64{1},
			want:    map[int64]bool{1: true},
			wantErr: false,
		},
		{
			name:    "用户无点赞记录",
			uid:     1002,
			biz:     "article",
			bizIds:  []int64{1, 2, 3},
			want:    map[int64]bool{1: false, 2: false, 3: false},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dao.BatchFindLikeInfo(ctx, tt.uid, tt.biz, tt.bizIds)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	// 验证数据库错误场景
	// t.Run("数据库错误", func(t *testing.T) {
	// 	// 创建错误表结构导致查询失败
	// 	db.Migrator().DropTable(&LikeInfo{})
	// 	_, err := dao.BatchFindLikeInfo(ctx, 1001, "article", []int64{1})
	// 	assert.Error(t, err)
	// })
}
