# Analysis Execution Log Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move analysis execution log admin read/search/share/report routes off generic 501 stubs.

**Architecture:** Add a narrow service surface for `project_analysis_plan_execution_log` backed by the existing GORM repository. Register a dedicated handler for get/search/html-report/share-token while leaving `test-run` stubbed until plan execution orchestration exists.

**Tech Stack:** Go, Gin, GORM, sqlite test DB.

---

### Task 1: Service Contract

**Files:**
- Create: `backend/internal/service/analysis_execution_log_service_test.go`
- Create: `backend/internal/service/analysis_execution_log_service.go`

- [ ] Write failing service tests for get validation, search defaults/filter normalization, and share token generation.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run AnalysisExecutionLog -count=1` and verify failure.
- [ ] Implement the minimal service types and methods.
- [ ] Re-run the service tests and verify pass.

### Task 2: Repository Queries

**Files:**
- Modify: `backend/internal/repository/analysis_execution_log_repository_test.go`
- Modify: `backend/internal/repository/review_log_repository.go`

- [ ] Write failing repository tests for search filters/pagination and share token update.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run AnalysisExecutionLog -count=1` and verify failure.
- [ ] Implement search and share-token update on `ReviewLogRepository`.
- [ ] Re-run repository tests and verify pass.

### Task 3: Handler And Router

**Files:**
- Create: `backend/internal/handler/analysis_execution_log_handler.go`
- Create: `backend/internal/handler/analysis_execution_log_handler_test.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_contract_test.go`
- Modify: `backend/cmd/server/main.go`

- [ ] Write failing handler/router tests for get/search/share/html-report.
- [ ] Run targeted handler/router tests and verify failure.
- [ ] Implement handler and wire routes/dependencies.
- [ ] Re-run targeted tests and verify pass.

### Task 4: Verification And Commit

- [ ] Run `gofmt` on touched Go files.
- [ ] Run focused package tests.
- [ ] Run non-network touched package suite.
- [ ] Run `git diff --check`.
- [ ] Commit the coherent slice.
