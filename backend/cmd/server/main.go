package main

import (
	"log"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/config"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	r := router.New()
	if err := r.Run(cfg.Server.Address()); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
