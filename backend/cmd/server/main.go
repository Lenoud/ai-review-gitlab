package main

import (
	"context"
	"log"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/config"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/database"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/handler"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/middleware"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/repository"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/router"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate database: %v", err)
	}
	if err := database.BootstrapAdmin(context.Background(), db, database.BootstrapAdminOptions{
		Username: cfg.Auth.AdminUsername,
		Password: cfg.Auth.AdminPassword,
		Nickname: "Administrator",
	}); err != nil {
		log.Fatalf("bootstrap admin: %v", err)
	}

	accessTTL, err := time.ParseDuration(cfg.Auth.AccessTokenTTL)
	if err != nil {
		log.Fatalf("parse auth.access_token_ttl: %v", err)
	}
	refreshTTL, err := time.ParseDuration(cfg.Auth.RefreshTokenTTL)
	if err != nil {
		log.Fatalf("parse auth.refresh_token_ttl: %v", err)
	}
	authSvc, err := service.NewAuthService(repository.NewUserRepository(db), service.AuthOptions{
		JWTSecret:       cfg.Auth.JWTSecret,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
		Issuer:          cfg.Auth.Issuer,
	})
	if err != nil {
		log.Fatalf("init auth service: %v", err)
	}

	authHandler := handler.NewAuthHandler(authSvc)
	r := router.New(router.Dependencies{
		AuthHandler:    authHandler,
		AuthMiddleware: middleware.JWTAuth(authSvc),
	})
	if err := r.Run(cfg.Server.Address()); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
