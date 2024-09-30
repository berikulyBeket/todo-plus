package main

import (
	"log"

	"github.com/berikulyBeket/todo-plus/config"
	"github.com/berikulyBeket/todo-plus/internal/app"
)

func main() {
	cfg, err := config.NewConfig("./config/config.yml")
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
