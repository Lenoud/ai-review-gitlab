# Analysis Plan Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move project analysis plan admin CRUD/search routes off generic 501 stubs.

**Architecture:** Reuse the existing `project_analysis_plan` schema/model and keep the established `handler -> service -> repository -> model` layering. This slice handles persisted plan configuration only; `/project-analysis-plan-execution-log/test-run` remains stubbed because execution scheduling/orchestration is separate scope.

**Tech Stack:** Go, Gin, GORM, sqlite repository tests, Testify.

---

## Scope

Implement:

- `POST /api/v1/admin/project-analysis-plan/create`
- `POST /api/v1/admin/project-analysis-plan/update`
- `GET /api/v1/admin/project-analysis-plan/get`
- `POST /api/v1/admin/project-analysis-plan/delete`
- `GET /api/v1/admin/project-analysis-plan/search`

## File Structure

- Create `backend/internal/service/project_analysis_plan_service.go`: validation, normalization, repository interface, page DTOs.
- Create `backend/internal/repository/project_analysis_plan_repository.go`: GORM persistence for CRUD/search.
- Create `backend/internal/handler/project_analysis_plan_handler.go`: request binding and response/error mapping.
- Modify `backend/internal/router/router.go`: inject handler and register routes.
- Modify `backend/cmd/server/main.go`: wire repository/service/handler.
- Modify `backend/internal/router/router_contract_test.go`: route contract expects real handler statuses.
- Add focused tests in service, repository, handler, and router packages.

---

### Task 1: Service Contract

- [x] Write failing service tests for create validation/defaults, update missing id, delete empty ids, search pagination normalization, and get validation.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run ProjectAnalysisPlan -count=1` from `backend` and verify it fails.
- [x] Implement the service DTOs, repository interface, normalization, and methods.
- [x] Run the same service tests and verify they pass.

### Task 2: Repository Persistence

- [x] Write failing repository tests with sqlite in-memory DB for create, update, get not found mapping, delete, and filtered search.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run ProjectAnalysisPlan -count=1` from `backend` and verify it fails.
- [x] Implement the GORM repository.
- [x] Run the repository tests and verify they pass.

### Task 3: Handler And Router

- [x] Write failing handler tests for create/get/search/delete and router contract expectations for real statuses.
- [x] Run targeted handler/router tests and verify they fail.
- [x] Implement handler, route registration, dependency injection, and server wiring.
- [x] Run targeted handler/router tests and verify they pass.

### Task 4: Verification And Commit

- [x] Run the touched package suite.
- [x] Run compile-only all-package verification.
- [x] Run `git diff --check`.
- [x] Commit the coherent slice with the project commit-message style.
