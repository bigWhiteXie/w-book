//go:build wireinject
// +build wireinject

package startup

import (
	"testing"

	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/internal/dao/cache"
	dao "codexie.com/w-book-article/internal/dao/db"
	"codexie.com/w-book-article/internal/handler"
	"codexie.com/w-book-article/internal/logic"
	"codexie.com/w-book-article/internal/repo"

	"codexie.com/w-book-common/kafka/producer"

	"codexie.com/w-book-article/internal/svc"
	"github.com/google/wire"
)

var AppSet = wire.NewSet(NewArticleApp)

var HandlerSet = wire.NewSet(handler.NewArticleHandler)

var LogicSet = wire.NewSet(logic.NewArticleLogic, logic.NewRankingLogic)

var SvcSet = wire.NewSet(svc.NewServiceContext)

var RepoSet = wire.NewSet(repo.NewAuthorRepository, repo.NewReaderRepository, repo.NewRankRepo)

var DaoSet = wire.NewSet(dao.NewAuthorDao, dao.NewReaderDao)

var CacheSet = wire.NewSet(cache.NewRankCacheRedis, cache.NewLocalArtTopCache, cache.NewArticleRedis)

var DbSet = wire.NewSet(InitGormDB, InitRedis, InitRedLock)

var MessageSet = wire.NewSet(InitKafkaClient, producer.NewKafkaProducer)

var RpcSet = wire.NewSet(InitInteractClient)

func NewApp(c config.Config, t *testing.T) (*App, error) {
	panic(wire.Build(
		AppSet,
		HandlerSet,
		LogicSet,
		SvcSet,
		RepoSet,
		DaoSet,
		CacheSet,
		DbSet,
		MessageSet,
		RpcSet,
	))
}
