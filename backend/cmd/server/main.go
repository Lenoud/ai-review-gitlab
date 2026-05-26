package main

import (
	"context"
	"log"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/config"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/database"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/handler"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/llm"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/middleware"
	platformgitlab "github.com/Lenoud/ai-review-gitlab/backend/internal/platform/gitlab"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/repository"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/router"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/worker"
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

	projectSvc := service.NewProjectService(repository.NewProjectRepository(db))
	gitLabSvc := service.NewGitLabService(platformgitlab.NewServiceAdapter(nil))
	reviewTaskSvc := service.NewReviewTaskService(repository.NewProjectRepository(db), repository.NewReviewTaskRepository(db), service.ReviewTaskOptions{})
	reviewLogSvc := service.NewReviewLogService(repository.NewReviewLogRepository(db))
	aiReviewTraceSvc := service.NewAIReviewTraceService(repository.NewAIReviewTraceRepository(db))
	openReportSvc := service.NewOpenReportService(repository.NewReviewLogRepository(db))
	llmModelSvc := service.NewLLMModelService(repository.NewLLMModelRepository(db), llm.NewOpenAICompatibleChecker(nil))
	workerCtx, stopWorker := context.WithCancel(context.Background())
	var workerRunner *worker.Runner
	if cfg.Worker.Enabled {
		idleInterval, err := time.ParseDuration(cfg.Worker.IdleInterval)
		if err != nil {
			log.Fatalf("parse worker.idle_interval: %v", err)
		}
		errorInterval, err := time.ParseDuration(cfg.Worker.ErrorInterval)
		if err != nil {
			log.Fatalf("parse worker.error_interval: %v", err)
		}
		reviewWorkerSvc := service.NewReviewWorkerService(
			reviewTaskSvc,
			repository.NewProjectRepository(db),
			repository.NewLLMModelRepository(db),
			platformgitlab.NewServiceAdapter(nil),
			llm.NewOpenAICompatibleClient(nil),
			repository.NewReviewLogRepository(db),
			repository.NewAIReviewTraceRepository(db),
			service.ReviewWorkerOptions{MaxInputTokens: cfg.Worker.MaxInputTokens},
		)
		workerRunner = worker.NewRunner(reviewWorkerSvc, worker.RunnerOptions{
			WorkerID:      cfg.Worker.ID,
			IdleInterval:  idleInterval,
			ErrorInterval: errorInterval,
		})
		workerRunner.Start(workerCtx)
		log.Printf("review worker started: id=%s", cfg.Worker.ID)
	}
	defer func() {
		stopWorker()
		if workerRunner != nil {
			workerRunner.Wait()
		}
	}()
	authHandler := handler.NewAuthHandler(authSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	projectGitLabHandler := handler.NewProjectGitLabHandler(gitLabSvc)
	llmModelHandler := handler.NewLLMModelHandler(llmModelSvc)
	reviewLogHandler := handler.NewReviewLogHandler(reviewLogSvc)
	aiReviewTraceHandler := handler.NewAIReviewTraceHandler(aiReviewTraceSvc)
	openReportHandler := handler.NewOpenReportHandler(openReportSvc)
	webhookHandler := handler.NewWebhookHandler(reviewTaskSvc)
	r := router.New(router.Dependencies{
		AuthHandler:          authHandler,
		ProjectHandler:       projectHandler,
		ProjectGitLabHandler: projectGitLabHandler,
		LLMModelHandler:      llmModelHandler,
		ReviewLogHandler:     reviewLogHandler,
		AIReviewTraceHandler: aiReviewTraceHandler,
		OpenReportHandler:    openReportHandler,
		WebhookHandler:       webhookHandler,
		AuthMiddleware:       middleware.JWTAuth(authSvc),
	})
	if err := r.Run(cfg.Server.Address()); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
