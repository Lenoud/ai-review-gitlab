# Project Template Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move project-template admin CRUD/list routes off generic 501 stubs.

**Architecture:** Follow the existing `handler -> service -> repository -> model` shape used by project analysis plans. This slice covers only `project_template`; `project_template_review_rule` remains stubbed for a separate increment.

**Tech Stack:** Go 1.22, Gin, GORM, sqlite in-memory repository tests, MySQL migrations.

---

### Task 1: Service Contract

**Files:**
- Create: `backend/internal/service/project_template_service.go`
- Test: `backend/internal/service/project_template_service_test.go`

- [x] Write failing service tests for create normalization, invalid create, update/get ID validation, delete ID cleanup, and list keyword trimming.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run ProjectTemplate -count=1` from `backend`; expect compile failure because service symbols do not exist.
- [x] Implement `ProjectTemplateService`, domain structs, errors, repository interface, input normalization, delete cleanup, reference guard, and full-list query.
- [x] Re-run the service test command; expect pass.

### Task 2: Model And Repository

**Files:**
- Create: `backend/internal/model/project_template.go`
- Create: `backend/internal/repository/project_template_repository.go`
- Test: `backend/internal/repository/project_template_repository_test.go`
- Modify: `backend/internal/database/database.go`
- Create: `backend/migrations/000008_project_template_schema.sql`

- [x] Write failing repository tests for create/update/find, not-found mapping, delete, and list filtering.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run ProjectTemplate -count=1` from `backend`; expect compile failure because repository/model symbols do not exist.
- [x] Implement model, repository, AutoMigrate registration, and SQL migration.
- [x] Re-run the repository test command; expect pass.

### Task 3: Handler

**Files:**
- Create: `backend/internal/handler/project_template_handler.go`
- Test: `backend/internal/handler/project_template_handler_test.go`

- [x] Write failing handler tests for create/get/list/delete happy path and not-found mapping.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/handler -run ProjectTemplate -count=1` from `backend`; expect compile failure because handler symbols do not exist.
- [x] Implement handler request parsing, response envelopes, query parsing, and error mapping.
- [x] Re-run the handler test command; expect pass.

### Task 4: Router And Server Wiring

**Files:**
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_contract_test.go`
- Modify: `backend/cmd/server/main.go`

- [x] Write failing router expectations that project-template CRUD/list routes are no longer generic 501 stubs.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/router -run 'TestAdminRoutesRequireAuthAndReturnNotImplementedWithDevToken|TestAdminProtectedRoutesReturnForbiddenWithoutPermission' -count=1` from `backend`; expect compile failure or mismatched status until router wiring exists.
- [x] Add `ProjectTemplateHandler` dependency, route registration, contract fake, and server DI wiring.
- [x] Re-run the router test command; expect pass.

### Task 5: Verification And Commit

**Files:**
- All files above

- [x] Run focused tests:
  `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service ./internal/repository ./internal/handler -run ProjectTemplate -count=1`
- [x] Run router contract tests:
  `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/router -run 'TestAdminRoutesRequireAuthAndReturnNotImplementedWithDevToken|TestAdminProtectedRoutesReturnForbiddenWithoutPermission' -count=1`
- [x] Run non-network package suite:
  `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/config ./internal/database ./internal/handler ./internal/middleware ./internal/pkg/response ./internal/repository ./internal/router ./internal/service ./internal/worker -count=1`
- [x] Run compile-only all-package check:
  `GOCACHE=/private/tmp/ai-revire-go-cache go test ./... -run '^TestDoesNotExist$'`
- [x] Run `git diff --check`.
- [ ] Commit coherent changes.
