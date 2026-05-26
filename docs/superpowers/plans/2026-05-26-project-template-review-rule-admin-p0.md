# Project Template Review Rule Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move project-template review-rule admin CRUD/list routes off generic 501 stubs.

**Architecture:** Follow the existing `handler -> service -> repository -> model` shape used by project-template CRUD. Store `globPatterns` as JSON in `project_template_review_rule`, validate referenced templates, preserve explicit `enabled=false`, and order rules by priority descending then creation time ascending.

**Tech Stack:** Go, Gin, GORM, SQLite tests, MySQL migrations.

---

### Task 1: Service Contract

**Files:**
- Create: `backend/internal/service/project_template_review_rule_service_test.go`
- Create: `backend/internal/service/project_template_review_rule_service.go`

- [x] Write failing tests for create normalization, invalid input, template missing, update template mismatch, list-by-template-id, get, and delete.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run ProjectTemplateReviewRule -count=1` from `backend` and confirm RED.
- [x] Implement minimal service types, errors, repository interface, normalization, and guards.
- [x] Re-run service tests and confirm GREEN.

### Task 2: Persistence

**Files:**
- Modify: `backend/internal/model/project_template.go`
- Create: `backend/internal/repository/project_template_review_rule_repository_test.go`
- Create: `backend/internal/repository/project_template_review_rule_repository.go`
- Modify: `backend/internal/database/database.go`
- Modify: `backend/migrations/000008_project_template_schema.sql`

- [x] Write failing repository tests for create/update/get/delete/list ordering/template existence/glob JSON mapping.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run ProjectTemplateReviewRule -count=1` from `backend` and confirm RED.
- [x] Add the model, repository, AutoMigrate entry, and SQL migration table.
- [x] Re-run repository tests and confirm GREEN.

### Task 3: Handler And Routes

**Files:**
- Create: `backend/internal/handler/project_template_review_rule_handler_test.go`
- Create: `backend/internal/handler/project_template_review_rule_handler.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_contract_test.go`
- Modify: `backend/cmd/server/main.go`

- [x] Write failing handler and router contract tests for create/update/get/delete/list-by-template-id.
- [x] Run targeted handler/router tests and confirm RED.
- [x] Add handler, dependency injection, and route registration; remove these routes from generic stubs.
- [x] Re-run targeted handler/router tests and confirm GREEN.

### Task 4: Verification And Review

**Files:**
- All touched files.

- [x] Run focused review-rule tests.
- [x] Run non-network backend suite used by prior automation runs.
- [x] Run compile-only `go test ./... -run '^TestDoesNotExist$'`.
- [x] Run `git diff --check`.
- [x] Request code review and fix Critical/Important findings.
- [x] Commit a coherent change if verification passes.
