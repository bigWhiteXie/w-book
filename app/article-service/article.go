package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-article/internal/config"
	"codexie.com/w-book-article/ioc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "/usr/local/go_project/w-book/app/article-service/etc/article.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	cron := cron.New()
	app, err := ioc.NewApp(cron, c, c.MySQLConf, c.RedisConf, c.KafkaConf)
	if err != nil {
		panic(err)
	}
	defer func() {
		app.Server.Stop()
		app.JobStarter.Stop()
	}()
	app.JobStarter.Start()
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	app.Server.Start()
}
