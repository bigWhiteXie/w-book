package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-reward/internal/config"
	"codexie.com/w-book-reward/ioc"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "/usr/local/go_project/w-book/app/reward-service/etc/reward.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	app, err := ioc.NewApp(c)
	if err != nil {
		panic(err)
	}
	defer func() {
		app.Server.Stop()
	}()
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	app.Server.Start()
}
