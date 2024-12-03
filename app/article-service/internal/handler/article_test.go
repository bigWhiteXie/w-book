package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/dao/cache"
	"codexie.com/w-book-article/internal/dao/db"
	"codexie.com/w-book-article/internal/domain"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-article/internal/svc"
	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/kafka/producer"
	"github.com/redis/go-redis/v9"
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
	cache  *redis.Client
	server *rest.Server
}

func (s *ArticleHandlerSuite) SetupSuite() {
	var configFile = flag.String("f", "/usr/local/go_project/w-book/app/article-service/etc/article.yaml", "the config file")
	var c config.Config
	conf.MustLoad(*configFile, &c)

	serviceContext := svc.NewServiceContext(c)
	gormDB := ioc.InitGormDB(c.MySQLConf)
	authorDao := db.NewAuthorDao(gormDB)
	client := ioc.InitRedis(c.RedisConf)
	articleCache := cache.NewArticleRedis(client)
	iAuthorRepository := repo.NewAuthorRepository(authorDao, articleCache)
	readerDao := db.NewReaderDao(gormDB)
	iReaderRepository := repo.NewReaderRepository(readerDao, articleCache)
	interactionClient := svc.CreateCodeRpcClient(c)
	p := producer.NewKafkaProducer(ioc.InitKafkaClient(c.KafkaConf))
	articleLogic := logic.NewArticleLogic(iAuthorRepository, iReaderRepository, interactionClient, p)
	localArtTopCache := cache.NewLocalArtTopCache()
	redisArtTopNCache := cache.NewRankCacheRedis(client)
	rankRepo := repo.NewRankRepo(localArtTopCache, redisArtTopNCache)
	redsync := ioc.InitRedLock(c.RedisConf)
	rankingLogic := logic.NewRankingLogic(iReaderRepository, rankRepo, redsync, interactionClient)
	articleHandler := NewArticleHandler(serviceContext, articleLogic, rankingLogic)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	RegisterHandlers(server, articleHandler)
	s.server = server
}

func (s *ArticleHandlerSuite) TearDownTest() {
	s.db.Exec("truncate table `published_article`")

	s.db.Exec("truncate table `article`")
	keys, _ := s.cache.Keys(context.Background(), "*").Result()
	for _, k := range keys {
		s.cache.Del(context.Background(), k)
	}
}

func (s *ArticleHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []testCase[float64]{
		{
			name: "创建新文章",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// check db
				var art db.Article
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
			name: "update other's article",
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

func (s *ArticleHandlerSuite) TestPublish() {
	t := s.T()
	testCases := []testCase[float64]{
		{
			name: "创建并发布",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				var art db.PublishedArticle
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				now := time.Now().UnixMilli() - 3600*1000
				assert.True(t, art.Utime > now)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, db.PublishedArticle{
					Id:       1,
					Title:    "my title",
					Content:  "my content",
					AuthorId: int64(123),
					Status:   uint8(domain.ArticlePublishedStatus),
				}, art)
				var pubArt db.PublishedArticle
				err = s.db.Where("id=?", 1).First(&pubArt).Error
				assert.NoError(t, err)
				now = time.Now().UnixMilli() - 3600*1000
				assert.True(t, pubArt.Utime > now)
				pubArt.Ctime = 0
				pubArt.Utime = 0
				assert.Equal(t, db.PublishedArticle{
					Id:       1,
					Title:    "my title",
					Content:  "my content",
					AuthorId: int64(123),
					Status:   uint8(domain.ArticlePublishedStatus),
				}, pubArt)
				bytes, err := s.cache.Get(context.Background(), "article:firstpage:123").Bytes()
				assert.Equal(t, 0, len(bytes))
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
				s.db.Create(&db.Article{
					Id:       2,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 123,
					Ctime:    now,
					Utime:    now,
				})
			},
			after: func(t *testing.T) {
				var art db.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				now := time.Now().UnixMilli() - 10*1000
				assert.True(t, art.Utime > now)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, db.Article{
					Id:       2,
					Title:    "new title",
					Content:  "new content",
					AuthorId: int64(123),
				}, art)
				var pubArt db.PublishedArticle
				err = s.db.Where("id=?", 2).First(&pubArt).Error
				assert.NoError(t, err)
				now = time.Now().UnixMilli() - 3*1000
				assert.True(t, pubArt.Utime > now)
				pubArt.Ctime = 0
				pubArt.Utime = 0
				assert.Equal(t, db.PublishedArticle{
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

func (s *ArticleHandlerSuite) TestPage() {
	t := s.T()
	testCases := []testCase[[]*domain.Article]{
		{
			name: "查询文章首页列表",
			before: func(t *testing.T) {
				var art = &db.Article{}
				now := time.Now().UnixMilli()
				art.AuthorId = 123
				art.Content = "my content"
				art.Title = "my title"
				art.Ctime = now
				art.Utime = now
				s.db.Create(art)
			},
			after: func(t *testing.T) {
				var arts = domain.ArticleArray{}
				// 校验缓存是否存在
				bytes, err := s.cache.Get(context.Background(), "article:firstpage:123").Bytes()
				assert.NoError(t, err)
				err = json.Unmarshal(bytes, &arts)
				articles := []*domain.Article(arts)
				assert.NoError(t, err)
				assert.Equal(t, "my title", articles[0].Title)

			},
			req: Article{
				Page: 1,
				Size: 1,
			},
			wantCode: http.StatusOK,
			wantRes: Result[[]*domain.Article]{
				Code: 200,
				Msg:  "ok",
				Data: []*domain.Article{&domain.Article{Id: 1, Title: "my title", Content: "my content"}},
			},
		},
		{
			name: "查询文章第二页列表",
			before: func(t *testing.T) {
				var art = &db.Article{}
				now := time.Now().UnixMilli()
				art.AuthorId = 123
				art.Content = "my content1"
				art.Title = "my title1"
				art.Ctime = now
				art.Utime = now
				s.db.Create(art)
				var art2 = &db.Article{}
				art2.AuthorId = 123
				art2.Content = "my content2"
				art2.Title = "my title2"
				art2.Ctime = now
				art2.Utime = now
				s.db.Create(art2)
			},
			after: func(t *testing.T) {

			},
			req: Article{
				Page: 2,
				Size: 1,
			},
			wantCode: http.StatusOK,
			wantRes: Result[[]*domain.Article]{
				Code: 200,
				Msg:  "ok",
				Data: []*domain.Article{&domain.Article{Title: "my title1", Content: "my content1"}},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/article/list?page=%d&size=%d", tc.req.Page, tc.req.Size), nil)
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, request)
			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Result[[]*domain.Article]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes.Data[0].Title, res.Data[0].Title)
			assert.Empty(t, res.Data[0].Content)
		})
	}
}

func (s *ArticleHandlerSuite) TestViewArticle() {
	t := s.T()
	testCases := []testCase[*domain.Article]{
		{
			name: "查看文章内容",
			before: func(t *testing.T) {
				s.TearDownTest()
				//准备一篇文章入库
				var art = &db.Article{}
				now := time.Now().UnixMilli()
				art.AuthorId = 123
				art.Content = "my content"
				art.Title = "my title"
				art.Ctime = now
				art.Utime = now
				s.db.Create(art)
			},
			after: func(t *testing.T) {

			},
			req: Article{
				Page:      1,
				Size:      1,
				Id:        1,
				isPublish: "false",
			},
			wantCode: http.StatusOK,
			wantRes: Result[*domain.Article]{
				Code: 200,
				Msg:  "ok",
				Data: &domain.Article{Id: 1, Title: "my title", Content: "my content"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/article/view?id=%d&isPublished=%s", tc.req.Id, tc.req.isPublish), nil)
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, request)
			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Result[*domain.Article]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes.Data.Id, res.Data.Id)
			assert.NotEmpty(t, res.Data.Content)
		})
	}
}

type Article struct {
	Id        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
	isPublish string
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,optional"`
}
type testCase[T any] struct {
	name     string
	before   func(t *testing.T)
	after    func(t *testing.T)
	req      Article
	wantCode int
	wantRes  Result[T]
}
