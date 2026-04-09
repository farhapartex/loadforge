package main

import (
	"flag"
	"log"

	"github.com/farhapartex/loadforge/web"
)

func main() {
	configPath := flag.String("config", "web.yml", "path to web server config file")
	flag.Parse()

	if err := web.Start(*configPath); err != nil {
		log.Fatal(err)
	}
}
