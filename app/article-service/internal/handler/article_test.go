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

	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/dao"
	"codexie.com/w-book-article/internal/domain"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-article/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"gorm.io/gorm"
)

func TestArticleGormHandler(t *testing.T) {
	suite.Run(t, &ArticleHandlerSuite{})
}

type ArticleHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	server *rest.Server
}

func (s *ArticleHandlerSuite) SetupSuite() {
	var configFile = flag.String("f", "/usr/local/go_project/w-book/app/article-service/etc/article.yaml", "the config file")
	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	serviceContext := svc.NewServiceContext(c)
	s.db = svc.CreteDbClient(c)
	articleLogic := logic.NewArticleLogic(repo.NewAuthorRepository(dao.NewAuthorDao(s.db)), repo.NewReaderRepository(dao.NewReaderDao(s.db)))
	articleHandler := NewArticleHandler(serviceContext, articleLogic)
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/article/edit",
		Handler: articleHandler.EditArticle,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/article/publish",
		Handler: articleHandler.publish,
	})
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "id", 123)
			ctx = context.WithValue(ctx, "sid", "dsadas")
			next(w, r.WithContext(ctx))
		}
	})

	s.server = server
}

func (s *ArticleHandlerSuite) TearDownTest() {
	s.db.Exec("truncate table `published_article`")

	s.db.Exec("truncate table `article`")
}

func (s *ArticleHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []testCase{
		{
			name: "创建新文章",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// check db
				var art dao.Article
				err := s.db.Where("author_id=?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Ctime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: int64(123),
				}, art)
			},
			req: Article{
				Title:   "my article",
				Content: "my article content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[float64]{
				Code: 200,
				Msg:  "ok",
				Data: 1,
			},
		},
		{
			name: "修改文章内容",
			before: func(t *testing.T) {
				s.db.Create(&dao.Article{
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
				var art dao.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 123)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
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
			name: "update other's article",
			before: func(t *testing.T) {
				s.db.Create(&dao.Article{
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
				var art dao.Article
				err := s.db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
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

func (s *ArticleHandlerSuite) TestPublish() {
	t := s.T()
	testCases := []testCase{
		{
			name: "创建并发布",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				var art dao.PublishedArticle
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				now := time.Now().UnixMilli() - 3600*1000
				assert.True(t, art.Utime > now)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.PublishedArticle{
					Id:       1,
					Title:    "my title",
					Content:  "my content",
					AuthorId: int64(123),
					Status:   uint8(domain.ArticlePublishedStatus),
				}, art)
				var pubArt dao.PublishedArticle
				err = s.db.Where("id=?", 1).First(&pubArt).Error
				assert.NoError(t, err)
				now = time.Now().UnixMilli() - 3600*1000
				assert.True(t, pubArt.Utime > now)
				pubArt.Ctime = 0
				pubArt.Utime = 0
				assert.Equal(t, dao.PublishedArticle{
					Id:       1,
					Title:    "my title",
					Content:  "my content",
					AuthorId: int64(123),
					Status:   uint8(domain.ArticlePublishedStatus),
				}, pubArt)
			},
			req: Article{
				Title:   "my title",
				Content: "my content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[float64]{
				Code: 200,
				Msg:  "ok",
				Data: 1,
			},
		},
		{
			name: "修改并发布",
			before: func(t *testing.T) {
				now := time.Now().UnixMilli()
				s.db.Create(&dao.Article{
					Id:       2,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 123,
					Ctime:    now,
					Utime:    now,
				})
			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				now := time.Now().UnixMilli() - 10*1000
				assert.True(t, art.Utime > now)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "new title",
					Content:  "new content",
					AuthorId: int64(123),
				}, art)
				var pubArt dao.PublishedArticle
				err = s.db.Where("id=?", 2).First(&pubArt).Error
				assert.NoError(t, err)
				now = time.Now().UnixMilli() - 3*1000
				assert.True(t, pubArt.Utime > now)
				pubArt.Ctime = 0
				pubArt.Utime = 0
				assert.Equal(t, dao.PublishedArticle{
					Id:       2,
					Title:    "new title",
					Content:  "new content",
					AuthorId: int64(123),
					Status:   uint8(domain.ArticlePublishedStatus),
				}, pubArt)
			},
			req: Article{
				Id:      2,
				Title:   "new title",
				Content: "new content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[float64]{
				Code: 200,
				Msg:  "ok",
				Data: 2,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			body, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			request, err := http.NewRequest(http.MethodPost, "/v1//article/publish", bytes.NewReader(body))
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

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,optional"`
}
type testCase struct {
	name     string
	before   func(t *testing.T)
	after    func(t *testing.T)
	req      Article
	wantCode int
	wantRes  Result[float64]
}
