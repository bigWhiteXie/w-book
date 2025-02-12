package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTest() (*gorm.DB, *redis.Client, *commentRepo) {
	// 初始化真实数据库连接
	dsn := "root:j3391111@tcp(127.0.0.1:3306)/w-book-interaction?charset=utf8mb4&parseTime=True&loc=Local"
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	gormDB.AutoMigrate(&db.Comment{})

	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "192.168.126.100:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// 初始化repo
	commentCache := cache.NewCommentRedisCache(redisClient)
	repo := NewCommentRepo(gormDB, commentCache)

	return gormDB, redisClient, repo.(*commentRepo)
}

func clearTestData(gormDB *gorm.DB) {
	// 清理测试数据
	gormDB.Exec("truncate table comment")
}

func TestCommentRepo_GetRootComments(t *testing.T) {
	gormDB, redisClient, repo := setupTest()
	defer redisClient.Close()
	defer clearTestData(gormDB)

	// 准备测试数据
	testComments := []*db.Comment{
		{Biz: "test", BizID: 1, Content: "comment 1", Score: 150, ParentID: sql.NullInt64{Int64: 0, Valid: false}},
		{Biz: "test", BizID: 1, Content: "comment 2", Score: 120, ParentID: sql.NullInt64{Int64: 0, Valid: false}},
		{Biz: "test", BizID: 1, Content: "comment 3", Score: 100, ParentID: sql.NullInt64{Int64: 0, Valid: false}},
	}
	if err := gormDB.Create(testComments).Error; err != nil {
		t.Fatalf("准备测试数据失败: %v", err)
	}

	tests := []struct {
		name        string
		biz         string
		bizID       int64
		offset      int
		size        int
		lastScore   int64
		mockCache   func()
		expectCount int
		expectError bool
		prepareData func() []*db.Comment
	}{
		{
			name:      "缓存命中返回缓存数据",
			biz:       "test",
			bizID:     1,
			offset:    0,
			size:      2,
			lastScore: 0,
			mockCache: func() {
				cached := []*domain.Comment{
					{ID: 100, Content: "cached comment"},
				}
				repo.cache.SetRootComments(context.Background(), "test", 1, cached, time.Hour)
			},
			expectCount: 1,
		},
		{
			name:        "缓存未命中从数据库获取",
			biz:         "test",
			bizID:       1,
			offset:      0,
			size:        2,
			lastScore:   200,
			mockCache:   nil,
			expectCount: 2,
			prepareData: func() []*db.Comment {
				return []*db.Comment{
					{Biz: "test", BizID: 1, RootID: 1, Content: "comment 1", Score: 180, ParentID: sql.NullInt64{Int64: 0, Valid: false}},
					{Biz: "test", BizID: 1, RootID: 2, Content: "comment 2", Score: 160, ParentID: sql.NullInt64{Int64: 0, Valid: false}},
					{Biz: "test", BizID: 1, RootID: 3, Content: "comment 3", Score: 140, ParentID: sql.NullInt64{Int64: 0, Valid: false}},
					{Biz: "test", BizID: 1, RootID: 4, Content: "comment 4", Score: 120, ParentID: sql.NullInt64{Int64: 0, Valid: false}},
					{Biz: "test", BizID: 1, RootID: 5, Content: "comment 5", Score: 100, ParentID: sql.NullInt64{Int64: 0, Valid: false}},
					{Biz: "test", BizID: 1, RootID: 1, Content: "child1 of comment 1", Score: 80, ParentID: sql.NullInt64{Int64: 1, Valid: true}, Ctime: time.Now().Add(-1 * time.Hour).UnixMilli()},
					{Biz: "test", BizID: 1, RootID: 1, Content: "child2 of comment 1", Score: 70, ParentID: sql.NullInt64{Int64: 1, Valid: true}, Ctime: time.Now().Add(-2 * time.Hour).UnixMilli()},
					{Biz: "test", BizID: 1, RootID: 1, Content: "child3 of comment 1", Score: 60, ParentID: sql.NullInt64{Int64: 1, Valid: true}, Ctime: time.Now().Add(-3 * time.Hour).UnixMilli()},
					{Biz: "test", BizID: 1, RootID: 1, Content: "child4 of comment 1", Score: 50, ParentID: sql.NullInt64{Int64: 1, Valid: true}, Ctime: time.Now().Add(-4 * time.Hour).UnixMilli()},
					{Biz: "test", BizID: 1, RootID: 1, Content: "child5 of comment 1", Score: 40, ParentID: sql.NullInt64{Int64: 1, Valid: true}, Ctime: time.Now().Add(-5 * time.Hour).UnixMilli()},
				}
			},
		},
		{
			name:        "分页查询带lastScore",
			biz:         "test",
			bizID:       1,
			offset:      0,
			size:        2,
			lastScore:   130,
			expectCount: 2, // 应返回score < 130的记录（120, 100）
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理测试数据并重新准备
			clearTestData(gormDB)
			var testData []*db.Comment
			if tt.prepareData != nil {
				testData = tt.prepareData()
			} else {
				testData = testComments // 默认使用原始测试数据
			}
			if err := gormDB.Create(testData).Error; err != nil {
				t.Fatalf("准备测试数据失败: %v", err)
			}

			// 清理缓存（确保缓存未命中）
			repo.cache.DelRootComments(context.Background(), tt.biz, tt.bizID)

			if tt.mockCache != nil {
				tt.mockCache()
			}

			comments, err := repo.GetRootComments(context.Background(), tt.biz, tt.bizID, tt.offset, tt.size, tt.lastScore)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, comments, tt.expectCount)
				// 验证排序是否正确
				if len(comments) > 1 {
					assert.GreaterOrEqual(t, comments[0].Score, comments[1].Score)
				}

				// 增加更详细的断言
				if tt.name == "缓存未命中从数据库获取" {
					// 验证返回的是score最高的两条
					assert.Equal(t, int64(180), comments[0].Score)
					assert.Equal(t, int64(160), comments[1].Score)
					// 验证总数匹配数据库记录
					var dbCount int64
					gormDB.Model(&db.Comment{}).Where("biz = ? AND biz_id = ?", "test", 1).Count(&dbCount)
					assert.Equal(t, 5, int(dbCount), "数据库应存在5条测试记录")
				}
			}
		})
	}
}

func TestCommentRepo_GetChildComments(t *testing.T) {
	gormDB, redisClient, repo := setupTest()
	defer redisClient.Close()
	defer clearTestData(gormDB)

	// 准备测试数据
	rootComment := &db.Comment{Biz: "test", BizID: 1, Content: "root comment"}
	if err := gormDB.Create(rootComment).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	testComments := []*db.Comment{
		{RootID: rootComment.ID, Content: "child 1", Ctime: now.Add(-2 * time.Hour).UnixMilli()},
		{RootID: rootComment.ID, Content: "child 2", Ctime: now.Add(-1 * time.Hour).UnixMilli()},
		{RootID: rootComment.ID, Content: "child 3", Ctime: now.UnixMilli()},
	}
	if err := gormDB.Create(testComments).Error; err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		rootID      int64
		lastCTime   int64
		offset      int
		size        int
		mockCache   func()
		expectCount int
		expectError bool
	}{
		{
			name:   "缓存命中返回缓存数据",
			rootID: rootComment.ID,
			offset: 0,
			size:   2,
			mockCache: func() {
				cached := []*domain.Comment{
					{ID: 100, Content: "cached child"},
				}
				repo.cache.SetChildComments(context.Background(), rootComment.ID, cached, time.Hour)
			},
			expectCount: 1,
		},
		{
			name:        "缓存未命中从数据库获取最新",
			rootID:      rootComment.ID,
			lastCTime:   0,
			offset:      0,
			size:        2,
			expectCount: 2,
		},
		{
			name:        "分页查询带lastCTime",
			rootID:      rootComment.ID,
			lastCTime:   now.Add(-1 * time.Hour).UnixMilli(),
			offset:      0,
			size:        2,
			expectCount: 1, // 应返回ctime < 指定时间的记录
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理缓存
			repo.cache.DelChildComments(context.Background(), tt.rootID)

			if tt.mockCache != nil {
				tt.mockCache()
			}

			comments, err := repo.GetChildComments(context.Background(), tt.rootID, tt.lastCTime, tt.offset, tt.size)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, comments, tt.expectCount)
				// 验证时间排序是否正确
				if len(comments) > 1 {
					assert.Greater(t, comments[0].Ctime.UnixMilli(), comments[1].Ctime.UnixMilli())
				}
			}
		})
	}
}

func TestCommentRepo_CreateComment(t *testing.T) {
	gormDB, redisClient, repo := setupTest()
	defer redisClient.Close()
	defer clearTestData(gormDB)

	t.Run("创建根评论", func(t *testing.T) {

		// 创建新评论
		comment, err := repo.CreateComment(context.Background(), "test", 1, 0, "new root comment", 0)
		assert.NoError(t, err)
		assert.NotNil(t, comment)

		var dbComment db.Comment
		gormDB.Model(&db.Comment{}).Where("biz = ? AND biz_id = ? order by id desc limit 1", "test", 1).Scan(&dbComment)
		assert.Equal(t, dbComment.Content, comment.Content)

	})

	t.Run("创建子评论并验证缓存失效", func(t *testing.T) {
		// 创建根评论
		root, err := repo.CreateComment(context.Background(), "test", 1, 0, "root comment", 0)
		assert.NoError(t, err)

		// 创建子评论
		child, err := repo.CreateComment(context.Background(), "test", 1, root.ID, "new child comment", 0)
		assert.NoError(t, err)
		assert.NotNil(t, child)

		// 验证子评论的rootId等于根评论的id
		var dbComment db.Comment
		gormDB.Model(&db.Comment{}).Where("biz = ? AND biz_id = ? order by id desc limit 1", "test", 1).Scan(&dbComment)
		assert.Equal(t, dbComment.Content, child.Content)
		assert.Equal(t, dbComment.RootID, root.ID)
	})
}
