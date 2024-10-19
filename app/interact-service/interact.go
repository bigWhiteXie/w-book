package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-interact/internal/config"
	"codexie.com/w-book-interact/ioc"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/article.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server, err := ioc.NewApp(c)
	if err != nil {
		panic(err)
	}
	defer server.Stop()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
