package main

import (
	"log"

	"github.com/mathbdw/book/config"
	"github.com/mathbdw/book/internal/app"
)

func main() {
	cfg, err := config.ReadConfigYML("config.yml")
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.RunPublisher(cfg)
}
