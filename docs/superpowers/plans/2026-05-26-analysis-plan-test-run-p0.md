# Analysis Plan Test Run P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move `POST /api/v1/admin/project-analysis-plan-execution-log/test-run` off the generic 501 stub.

**Architecture:** Extend the existing analysis execution log service with a synchronous manual runner that validates project/prompt input, calls the default LLM model through the existing chat client interface, records a success or failure execution log, and returns the persisted log. Reuse the existing `ReviewLogRepository` for analysis execution persistence and the existing report rendering from saved logs; leave scheduled execution and outbound IM notification for separate slices.

**Tech Stack:** Go, Gin, GORM, sqlite repository tests, Testify, existing OpenAI-compatible LLM client.

---

### Task 1: Service Manual Test Run

**Files:**
- Modify: `backend/internal/service/analysis_execution_log_service.go`
- Modify: `backend/internal/service/analysis_execution_log_service_test.go`

- [ ] Write failing service tests for successful test-run recording, failure recording when LLM fails, and input validation.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run AnalysisExecutionLog -count=1` and verify failure.
- [ ] Implement minimal manual runner dependencies, input DTO, validation, LLM call, and success/failure recording.
- [ ] Re-run service tests and verify pass.

### Task 2: Repository Create Support

**Files:**
- Modify: `backend/internal/repository/review_log_repository.go`
- Modify: `backend/internal/repository/analysis_execution_log_repository_test.go`

- [ ] Write failing repository test for creating an analysis execution log with result and error fields.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run AnalysisExecutionLog -count=1` and verify failure.
- [ ] Implement `CreateAnalysisExecution` and map all relevant fields back to service DTOs.
- [ ] Re-run repository tests and verify pass.

### Task 3: Handler And Router Wiring

**Files:**
- Modify: `backend/internal/handler/analysis_execution_log_handler.go`
- Modify: `backend/internal/handler/analysis_execution_log_handler_test.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_contract_test.go`
- Modify: `backend/cmd/server/main.go`

- [ ] Write failing handler/router tests for `/test-run` request binding and route status.
- [ ] Run targeted handler/router tests and verify failure.
- [ ] Implement handler method, route registration, contract fake method, and server dependencies.
- [ ] Re-run targeted tests and verify pass.

### Task 4: Verification And Commit

- [ ] Run `gofmt` on touched Go files.
- [ ] Run focused analysis execution log tests.
- [ ] Run non-network package suite.
- [ ] Run compile-only all-package verification.
- [ ] Run `git diff --check` and staged diff check.
- [ ] Commit the coherent slice with the project commit-message style.
