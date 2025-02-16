package logic

import (
	"context"
	"flag"
	"testing"
	"time"

	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/kafka/producer"
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/types"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/conf"
	"gorm.io/gorm"
)

var saramaClient sarama.Client

func setupRealEnv() (*gorm.DB, *redis.Client, producer.Producer) {
	var configFile = flag.String("f", "/usr/local/go_project/w-book/app/interact-service/etc/interact-api.yaml", "the config file")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	client := ioc.InitRedis(c.RedisConf)
	db := ioc.InitGormDB(c.MySQLConf)

	saramaClient = ioc.InitKafkaClient(c.KafkaConf)
	producerProducer := producer.NewKafkaProducer(saramaClient)

	return db, client, producerProducer
}

func clearTestData(gormDB *gorm.DB, redisClient *redis.Client) {
	// 清理数据库测试数据
	gormDB.Exec("truncate table comment")
	gormDB.Exec("truncate table like_info")

	// 清理Redis测试数据
	redisClient.FlushDB(context.Background())
}

func TestCommentLogic_AddComment(t *testing.T) {
	gormDB, redisClient, kafkaProducer := setupRealEnv()
	defer clearTestData(gormDB, redisClient)

	// 初始化依赖组件
	commentRepo := repo.NewCommentRepo(gormDB, cache.NewCommentRedisCache(redisClient))
	logic := NewCommentLogic(commentRepo, kafkaProducer, redisClient)

	t.Run("创建根评论", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "id", 123)
		req := &types.AddCommentReq{
			Biz:     "test",
			BizID:   1,
			Content: "测试根评论",
		}

		comment, err := logic.AddComment(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, comment)
		assert.Equal(t, int64(0), comment.ParentID)
		assert.Equal(t, comment.ID, comment.RootID)

		// 验证数据库记录
		var dbComment db.Comment
		gormDB.Where("id = ?", comment.ID).First(&dbComment)
		assert.Equal(t, "测试根评论", dbComment.Content)
	})

	t.Run("创建子评论", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "id", 789)
		// 先创建根评论
		rootReq := &types.AddCommentReq{
			Biz:     "test",
			BizID:   1,
			Content: "父评论",
		}
		rootComment, _ := logic.AddComment(ctx, rootReq)

		// 创建子评论
		ctx = context.WithValue(context.Background(), "id", 456)
		req := &types.AddCommentReq{
			Biz:      "test",
			BizID:    1,
			ParentID: rootComment.ID,
			Content:  "子评论",
		}

		comment, err := logic.AddComment(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, rootComment.ID, comment.RootID)
		assert.Equal(t, rootComment.ID, comment.ParentID)

		// 验证数据库记录
		var dbComment db.Comment
		gormDB.Where("id = ?", comment.ID).First(&dbComment)
		assert.Equal(t, "子评论", dbComment.Content)
		assert.Equal(t, rootComment.ID, dbComment.RootID)
	})
}

func TestCommentLogic_GetRootComments(t *testing.T) {
	gormDB, redisClient, kafkaProducer := setupRealEnv()
	clearTestData(gormDB, redisClient)
	defer clearTestData(gormDB, redisClient)

	// 初始化依赖组件
	commentRepo := repo.NewCommentRepo(gormDB, cache.NewCommentRedisCache(redisClient))
	logic := NewCommentLogic(commentRepo, kafkaProducer, redisClient)

	// 准备测试数据
	ctx := context.WithValue(context.Background(), "id", 1001)
	root1, _ := logic.AddComment(ctx, &types.AddCommentReq{Biz: "test", BizID: 1, Content: "根评论1"})
	root2, _ := logic.AddComment(ctx, &types.AddCommentReq{Biz: "test", BizID: 1, Content: "根评论2"})

	// 添加子评论
	logic.AddComment(ctx, &types.AddCommentReq{Biz: "test", BizID: 1, ParentID: root1.ID, Content: "子评论1"})
	logic.AddComment(ctx, &types.AddCommentReq{Biz: "test", BizID: 1, ParentID: root1.ID, Content: "子评论2"})

	t.Run("获取带子评论的根评论列表", func(t *testing.T) {
		req := &types.GetRootCommentsReq{
			Biz:   "test",
			BizID: 1,
			Size:  10,
		}

		comments, err := logic.GetRootComments(ctx, req)
		assert.NoError(t, err)
		assert.Len(t, comments, 2)

		// 验证子评论加载
		for _, c := range comments {
			if c.ID == root1.ID {
				assert.Len(t, c.Childs, 2)
			}
		}
	})

	t.Run("验证点赞状态", func(t *testing.T) {
		// 给root2点赞
		logic.LikeComment(ctx, root2.ID, true)

		req := &types.GetRootCommentsReq{
			Biz:   "test",
			BizID: 1,
			Size:  10,
		}

		comments, _ := logic.GetRootComments(ctx, req)
		for _, c := range comments {
			if c.ID == root2.ID {
				assert.True(t, c.IsLike)
			}
		}
	})
}

func TestCommentLogic_LikeComment(t *testing.T) {
	gormDB, redisClient, kafkaProducer := setupRealEnv()
	clearTestData(gormDB, redisClient)
	// defer clearTestData(gormDB, redisClient)

	commentRepo := repo.NewCommentRepo(gormDB, cache.NewCommentRedisCache(redisClient))
	logic := NewCommentLogic(commentRepo, kafkaProducer, redisClient)

	// 创建测试评论
	ctx := context.WithValue(context.Background(), "id", 2001)
	comment, _ := logic.AddComment(ctx, &types.AddCommentReq{Biz: "test", BizID: 1, Content: "点赞测试评论"})

	t.Run("正常点赞", func(t *testing.T) {
		err := logic.LikeComment(ctx, comment.ID, true)
		commentRepo.HandleLikeCommentEvent(ctx, comment.ID, true, 2001)
		assert.NoError(t, err)

		// 验证数据库
		var likeInfo db.LikeInfo
		gormDB.Where("biz = ? AND biz_id = ?", domain.CommentBiz, comment.ID).First(&likeInfo)
		assert.Equal(t, uint8(1), likeInfo.Status)

		// 验证缓存
		status, _ := commentRepo.GetLikeStatusByUid(ctx, 2001, []int64{comment.ID})
		assert.True(t, status[comment.ID])
	})

	t.Run("取消点赞", func(t *testing.T) {
		err := logic.LikeComment(ctx, comment.ID, false)
		assert.NoError(t, err)
		commentRepo.HandleLikeCommentEvent(ctx, comment.ID, false, 2001)
		// 验证数据库
		var likeInfo db.LikeInfo
		time.Sleep(2 * time.Second)
		gormDB.Where("biz = ? AND biz_id = ?", domain.CommentBiz, comment.ID).First(&likeInfo)
		assert.Equal(t, uint8(0), likeInfo.Status)

		// 验证缓存
		status, _ := commentRepo.GetLikeStatusByUid(ctx, 2001, []int64{comment.ID})
		assert.False(t, status[comment.ID])
	})
}

func TestCommentLogic_DeleteComment(t *testing.T) {
	gormDB, redisClient, kafkaProducer := setupRealEnv()
	defer clearTestData(gormDB, redisClient)

	commentRepo := repo.NewCommentRepo(gormDB, cache.NewCommentRedisCache(redisClient))
	logic := NewCommentLogic(commentRepo, kafkaProducer, redisClient)
	rootCmt, _ := logic.AddComment(context.WithValue(context.Background(), "id", 3001), &types.AddCommentReq{Biz: "test", BizID: 1, Content: "根评论"})

	t.Run("删除自己的评论", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "id", 3001)
		comment, _ := logic.AddComment(ctx, &types.AddCommentReq{Biz: "test", BizID: 1, Content: "待删除评论", ParentID: rootCmt.ID})

		err := logic.DeleteComment(ctx, comment.ID)
		assert.NoError(t, err)

		// 验证数据库软删除
		var dbComment db.Comment
		gormDB.Unscoped().Where("id = ?", comment.ID).First(&dbComment)
		assert.NotZero(t, dbComment)
	})

	t.Run("删除他人评论应报错", func(t *testing.T) {
		ctx1 := context.WithValue(context.Background(), "id", 4001)
		comment, _ := logic.AddComment(ctx1, &types.AddCommentReq{Biz: "test", BizID: 1, Content: "他人评论"})

		ctx2 := context.WithValue(context.Background(), "id", 4002)
		err := logic.DeleteComment(ctx2, comment.ID)
		assert.Error(t, err)
	})
}
