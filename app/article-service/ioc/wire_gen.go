// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/dao/cache"
	"codexie.com/w-book-article/internal/dao/db"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-article/internal/job"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/repo"
	"codexie.com/w-book-article/internal/svc"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/rest"
)

// Injectors from wire.go:

func NewApp(c config.Config) (*rest.Server, error) {
	serviceContext := svc.NewServiceContext(c)
	gormDB := svc.CreteDbClient(c)
	authorDao := db.NewAuthorDao(gormDB)
	client := svc.CreateRedisClient(c)
	articleCache := cache.NewArticleRedis(client)
	iAuthorRepository := repo.NewAuthorRepository(authorDao, articleCache)
	readerDao := db.NewReaderDao(gormDB)
	iReaderRepository := repo.NewReaderRepository(readerDao, articleCache)
	interactionClient := svc.CreateCodeRpcClient(c)
	producer := svc.CreateKafkaProducer(c)
	articleLogic := logic.NewArticleLogic(iAuthorRepository, iReaderRepository, interactionClient, producer)
	localArtTopCache := cache.NewLocalArtTopCache()
	redisArtTopNCache := cache.NewRankCacheRedis(client)
	rankRepo := repo.NewRankRepo(localArtTopCache, redisArtTopNCache)
	redsync := svc.CreateRedSync(c)
	rankingLogic := logic.NewRankingLogic(iReaderRepository, rankRepo, redsync, interactionClient)
	articleHandler := handler.NewArticleHandler(serviceContext, articleLogic, rankingLogic)
	rankingJob := job.NewRankingJob(rankingLogic)
	jobBuilder := job.InitJobBuilder(rankingJob)
	server := NewServer(c, articleHandler, client, jobBuilder)
	return server, nil
}

// wire.go:

var ServerSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewArticleHandler)

var LogicSet = wire.NewSet(logic.NewArticleLogic, logic.NewRankingLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewAuthorRepository, repo.NewReaderRepository, repo.NewRankRepo)

var DaoSet = wire.NewSet(db.NewAuthorDao, db.NewReaderDao)

var CacheSet = wire.NewSet(cache.NewRankCacheRedis, cache.NewLocalArtTopCache, cache.NewArticleRedis)

var DbSet = wire.NewSet(svc.CreteDbClient, svc.CreateRedisClient, svc.CreateRedSync)

var MessageSet = wire.NewSet(svc.CreateKafkaProducer)

var RpcSet = wire.NewSet(svc.CreateCodeRpcClient)

var JobSet = wire.NewSet(job.InitJobBuilder, job.NewRankingJob)
