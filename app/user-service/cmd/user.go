package main

import (
	"flag"
	"fmt"

	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/ioc"
	_ "github.com/apache/skywalking-go"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "/usr/local/go_project/w-book/app/user-service/etc/user.yaml", "the config file")

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
