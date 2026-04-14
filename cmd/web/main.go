package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/farhapartex/loadforge/web"
)

func main() {
	defaultConfig := defaultConfigPath()
	configPath := flag.String("config", defaultConfig, "path to web server config file")
	flag.Parse()

	if err := web.Start(*configPath); err != nil {
		log.Fatal(err)
	}
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "web.yml"
	}
	return filepath.Join(home, ".loadforge", "web.yml")
}
