package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
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
	var configFile = flag.String("f", "/usr/local/go_project/w-book/app/interact-service/etc/interact-api.yaml", "the config file")
	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	serviceContext := svc.NewServiceContext(c)
	client := svc.CreateRedisClient(c)
	rs := svc.CreateRedSync(c)
	interactCache := cache.NewInteractRedis(client, rs)
	gormDB := svc.CreteDbClient(c)
	iLikeInfoRepository := repo.NewLikeInfoRepository(interactCache, gormDB)
	interactDao := db.NewInteractDao(gormDB)
	recordDao := db.NewRecordDao(gormDB)
	localCache := cache.NewBigCacheResourceCache()
	iInteractRepo := repo.NewInteractRepository(interactDao, recordDao, interactCache, localCache)
	collectionDao := db.NewCollectionDao(gormDB)
	iCollectRepository := repo.NewCollectRepository(interactCache, collectionDao)
	interactLogic := logic.NewInteractLogic(iLikeInfoRepository, iInteractRepo, iCollectRepository)
	interactHandler := NewInteractHandler(serviceContext, interactLogic)
	s.db = gormDB
	s.cache = client
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/resource/like",
		Handler: interactHandler.LikeResource,
	}, rest.WithTimeout(1000*time.Second))
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/resource/collection",
		Handler: interactHandler.OperateCollection,
	}, rest.WithTimeout(1000*time.Second))
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/resource/collect",
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
	testCases := []testCase[int, *types.OpResourceReq]{
		{
			name: "点赞未缓存数据",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// check db
				var like db.LikeInfo
				var interact db.Interaction
				err := s.db.Where("id=?", 1).First(&like).Error
				assert.NoError(t, err)
				assert.True(t, like.Uid == 123)
				assert.True(t, like.Status == 1)
				assert.True(t, like.Ctime > 0)

				s.db.Where("biz=? and biz_id=?", "article", 1).First(&interact)
				assert.True(t, interact.LikeCnt == 1)
			},
			req: &types.OpResourceReq{
				Biz:    "article",
				BizId:  1,
				Action: 1,
			},
			wantCode: http.StatusOK,
			wantRes: Result[int]{
				Code: 200,
				Msg:  "ok",
			},
		},
		{
			name: "取消点赞",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				// check db
				var like db.LikeInfo
				var interact db.Interaction
				err := s.db.Where("id=?", 1).First(&like).Error
				assert.NoError(t, err)
				assert.True(t, like.Uid == 123)
				assert.True(t, like.Status == 0)
				assert.True(t, like.Ctime > 0)

				s.db.Where("biz=? and biz_id=?", "article", 1).First(&interact)
				assert.True(t, interact.LikeCnt == 0)
			},
			req: &types.OpResourceReq{
				Biz:    "article",
				BizId:  1,
				Action: 0,
			},
			wantCode: http.StatusOK,
			wantRes: Result[int]{
				Code: 200,
				Msg:  "ok",
			},
		},
		{
			name: "点赞缓存数据",
			before: func(t *testing.T) {
				cntMap := map[string]string{
					domain.Like:    "0",
					domain.Collect: "0",
					domain.Read:    "0",
				}
				key := fmt.Sprintf("cnt:%s:%d", "article", 1)
				err := s.cache.HSet(context.Background(), key, cntMap).Err()
				assert.True(t, err == nil)
			},
			after: func(t *testing.T) {
				// check db
				var like db.LikeInfo
				var interact db.Interaction
				key := fmt.Sprintf("cnt:%s:%d", "article", 1)
				cntMap, _ := s.cache.HGetAll(context.Background(), key).Result()
				likeCnt, _ := strconv.Atoi(cntMap[domain.Like])
				assert.True(t, likeCnt == 1)

				err := s.db.Where("id=?", 1).First(&like).Error
				assert.NoError(t, err)
				assert.True(t, like.Uid == 123)
				assert.True(t, like.Status == 1)
				assert.True(t, like.Ctime > 0)
				s.db.Where("biz=? and biz_id=?", "article", 1).First(&interact)
				assert.True(t, interact.LikeCnt == 1)
			},
			req: &types.OpResourceReq{
				Biz:    "article",
				BizId:  1,
				Action: 1,
			},
			wantCode: http.StatusOK,
			wantRes: Result[int]{
				Code: 200,
				Msg:  "ok",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			body, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/resource/like", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, request)
			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Result[int]
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
