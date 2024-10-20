package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/svc"
	"codexie.com/w-book-interact/internal/types"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"gorm.io/gorm"
)

func TestArticleGormHandler(t *testing.T) {
	suite.Run(t, &InteractHandlerSuite{})
}

type InteractHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	cache  *redis.Client
	server *rest.Server
}

func (s *InteractHandlerSuite) SetupSuite() {
	var configFile = flag.String("f", "/usr/local/go_project/w-book/app/article-service/etc/article.yaml", "the config file")
	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	serviceContext := svc.NewServiceContext(c)
	client := svc.CreateRedisClient(c)
	interactCache := cache.NewInteractRedis(client)
	gormDB := svc.CreteDbClient(c)
	iLikeInfoRepository := repo.NewLikeInfoRepository(interactCache, gormDB)
	interactDao := db.NewInteractDao(gormDB)
	iInteractRepo := repo.NewInteractRepository(interactDao, interactCache)
	collectionDao := db.NewCollectionDao(gormDB)
	iCollectRepository := repo.NewCollectRepository(interactCache, collectionDao)
	interactLogic := logic.NewInteractLogic(iLikeInfoRepository, iInteractRepo, iCollectRepository)
	interactHandler := NewInteractHandler(serviceContext, interactLogic)
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/interact/like",
		Handler: interactHandler.LikeResource,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/interact/collection",
		Handler: interactHandler.OperateCollection,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/interact/collect",
		Handler: interactHandler.OperateCollectionItem,
	}, rest.WithTimeout(1000*time.Second))

	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "id", 123)
			ctx = context.WithValue(ctx, "sid", "dsadas")
			next(w, r.WithContext(ctx))
		}
	})

	s.server = server
}

func (s *InteractHandlerSuite) TearDownTest() {
	s.db.Exec("truncate table `like_info`")
	s.db.Exec("truncate table `interaction`")
	s.db.Exec("truncate table `collection`")

	keys, _ := s.cache.Keys(context.Background(), "*").Result()
	for _, k := range keys {
		s.cache.Del(context.Background(), k)
	}
}

func (s *InteractHandlerSuite) TestLike() {
	t := s.T()
	testCases := []testCase[float64, *types.LikeResourceReq]{
		{
			name: "点赞未缓存数据",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// check db
				var like db.LikeInfo
				var interact db.Interaction
				err := s.db.Where("author_id=?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Ctime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, db.Article{
					Id:       1,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: int64(123),
				}, art)
			},
			req: &types.LikeResourceReq{
				Biz:    "article",
				BizId:  1,
				Action: 1,
			},
			wantCode: http.StatusOK,
			wantRes: Result[float64]{
				Code: 200,
				Msg:  "ok",
				Data: 1,
			},
		},
		{
			name: "点赞已缓存数据",
			before: func(t *testing.T) {
				s.db.Create(&db.Article{
					Id:       2,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 123,
					Ctime:    123,
					Utime:    123,
				})
			},
			after: func(t *testing.T) {
				// check db
				var art db.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 123)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, db.Article{
					Id:       2,
					Title:    "new article",
					Content:  "new article content",
					AuthorId: int64(123),
				}, art)
			},
			req: Article{
				Id:      2,
				Title:   "new article",
				Content: "new article content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[float64]{
				Code: 200,
				Msg:  "ok",
				Data: 2,
			},
		},
		{
			name: "取消点赞未缓存数据",
			before: func(t *testing.T) {
				s.db.Create(&db.Article{
					Id:       3,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 456,
					Ctime:    123,
					Utime:    456,
				})
			},
			after: func(t *testing.T) {
				// check db
				var art db.Article
				err := s.db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, db.Article{
					Id:       3,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: int64(456),
					Ctime:    int64(123),
					Utime:    int64(456),
				}, art)
			},
			req: Article{
				Id:      3,
				Title:   "new article",
				Content: "new article content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[float64]{
				Code: 500,
				Msg:  "系统异常",
				Data: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			body, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/article/edit", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, request)
			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Result[float64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,optional"`
}
type testCase[T any, R any] struct {
	name     string
	before   func(t *testing.T)
	after    func(t *testing.T)
	req      R
	wantCode int
	wantRes  Result[T]
}
