package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-interact/api/pb/interact"
	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/logic"
	"codexie.com/w-book-interact/internal/repo"
	"codexie.com/w-book-interact/internal/server"
	"codexie.com/w-book-interact/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "/usr/local/go_project/w-book/app/interact-service/etc/interact-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	s := zrpc.MustNewServer(c.Grpc, func(grpcServer *grpc.Server) {
		interact.RegisterInteractionServer(grpcServer, NewServer(c))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.Grpc.ListenOn)
	s.Start()
}

func NewServer(c config.Config) *server.InteractionServer {
	serviceContext := svc.NewServiceContext(c)
	client := svc.CreateRedisClient(c)
	interactCache := cache.NewInteractRedis(client)
	gormDB := svc.CreteDbClient(c)
	iLikeInfoRepository := repo.NewLikeInfoRepository(interactCache, gormDB)
	interactDao := db.NewInteractDao(gormDB)
	recordDao := db.NewRecordDao(gormDB)

	iInteractRepo := repo.NewInteractRepository(interactDao, recordDao, interactCache)
	collectionDao := db.NewCollectionDao(gormDB)
	iCollectRepository := repo.NewCollectRepository(interactCache, collectionDao)
	interactLogic := logic.NewInteractLogic(iLikeInfoRepository, iInteractRepo, iCollectRepository)

	return server.NewInteractionServer(serviceContext, interactLogic)
}
