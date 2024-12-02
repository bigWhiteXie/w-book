//go:build wireinject
// +build wireinject

package ioc

import (
	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/dao/cache"
	dao "codexie.com/w-book-article/internal/dao/db"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-article/internal/job"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/repo"

	"codexie.com/w-book-article/internal/svc"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/rest"
)

var ServerSet = wire.NewSet(NewServer)

var HandlerSet = wire.NewSet(handler.NewArticleHandler)

var LogicSet = wire.NewSet(logic.NewArticleLogic, logic.NewRankingLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewAuthorRepository, repo.NewReaderRepository, repo.NewRankRepo)

var DaoSet = wire.NewSet(dao.NewAuthorDao, dao.NewReaderDao)

var CacheSet = wire.NewSet(cache.NewRankCacheRedis, cache.NewLocalArtTopCache, cache.NewArticleRedis)

var DbSet = wire.NewSet(svc.CreteDbClient, svc.CreateRedisClient, svc.CreateRedSync)

var MessageSet = wire.NewSet(svc.CreateKafkaProducer)

var RpcSet = wire.NewSet(svc.CreateCodeRpcClient)

var JobSet = wire.NewSet(job.InitJobBuilder, job.NewRankingJob)

func NewApp(c config.Config) (*rest.Server, error) {
	panic(wire.Build(
		ServerSet,
		HandlerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DaoSet,
		CacheSet,
		DbSet,
		MessageSet,
		RpcSet,
		JobSet,
	))
}
