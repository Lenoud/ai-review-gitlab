# Stats and Sys Log Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move `GET /api/v1/admin/stats`, `GET /api/v1/admin/member/commit-summary`, `GET /api/v1/admin/sys-log/get`, and `GET /api/v1/admin/sys-log/search` off generic 501 stubs.

**Architecture:** Follow the existing `handler -> service -> repository -> model` shape. Stats and member commit summary aggregate existing push and merge-request review logs. Sys-log adds a small persisted `sys_log` table and read/search APIs only; writing logs from middleware/workers remains separate scope.

**Tech Stack:** Go, Gin, GORM, MySQL migrations, sqlite-backed repository tests.

---

### Task 1: Plan and Contract Tests

**Files:**
- Create: `docs/superpowers/plans/2026-05-26-stats-sys-log-admin-p0.md`
- Modify: `backend/internal/router/router_contract_test.go`

- [x] **Step 1: Write this plan**
- [ ] **Step 2: Update router contract expectations**
  Remove stats/member/sys-log routes from the generic 501 list after handlers are wired.
- [ ] **Step 3: Run router contract test and confirm RED**
  Run: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/router -run TestAdminRoutesRequireAuthAndReturnNotImplementedWithDevToken -count=1`
  Expected: failure while routes are still generic 501 or missing handler dependencies.

### Task 2: Stats and Member Aggregates

**Files:**
- Create: `backend/internal/service/stats_service.go`
- Create: `backend/internal/service/stats_service_test.go`
- Create: `backend/internal/repository/stats_repository.go`
- Create: `backend/internal/repository/stats_repository_test.go`
- Create: `backend/internal/handler/stats_handler.go`
- Create: `backend/internal/handler/stats_handler_test.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Write service RED tests**
  Cover invalid time ranges, combined push/MR totals, weighted average scores, project/author/code-change summaries, member project counts, daily buckets, and paging.
- [ ] **Step 2: Run service tests and confirm RED**
  Run: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run Stats -count=1`
- [ ] **Step 3: Write repository RED tests**
  Seed review logs and assert aggregate results match the Java reference behavior.
- [ ] **Step 4: Run repository tests and confirm RED**
  Run: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run Stats -count=1`
- [ ] **Step 5: Implement minimal service/repository code**
  Add query normalization, time validation, aggregate DTOs, and read-only GORM aggregation.
- [ ] **Step 6: Add handler RED tests and implementation**
  Validate required `startTime`/`endTime`, bad ranges, success envelopes, and paged member summary response.

### Task 3: Sys Log Read/Search

**Files:**
- Create: `backend/internal/model/sys_log.go`
- Create: `backend/internal/repository/sys_log_repository.go`
- Create: `backend/internal/repository/sys_log_repository_test.go`
- Create: `backend/internal/service/sys_log_service.go`
- Create: `backend/internal/service/sys_log_service_test.go`
- Create: `backend/internal/handler/sys_log_handler.go`
- Create: `backend/internal/handler/sys_log_handler_test.go`
- Create: `backend/migrations/000014_sys_log_schema.sql`
- Modify: `backend/internal/database/database.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Write sys-log service/repository RED tests**
  Cover get not found, filtering by level/module/action/message/time, paging, and timestamp mapping.
- [ ] **Step 2: Run tests and confirm RED**
  Run: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service ./internal/repository -run SysLog -count=1`
- [ ] **Step 3: Implement model, migration, repository, and service**
  Keep this read/search only; no app-wide log writing yet.
- [ ] **Step 4: Add handler RED tests and implementation**
  Cover `id` validation, not found mapping, and search filters.

### Task 4: Wiring, Verification, Review, Commit

**Files:**
- Modify: `backend/internal/router/router.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/internal/router/router_contract_test.go`

- [ ] **Step 1: Wire router dependencies and remove implemented routes from stubs**
- [ ] **Step 2: Run focused tests**
  Run: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service ./internal/repository ./internal/handler ./internal/router -run 'Stats|SysLog|AdminRoutes' -count=1`
- [ ] **Step 3: Run non-network package suite**
  Run: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/config ./internal/database ./internal/handler ./internal/middleware ./internal/pkg/response ./internal/repository ./internal/router ./internal/service ./internal/worker -count=1`
- [ ] **Step 4: Run compile-only all-package check**
  Run: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./... -run '^TestDoesNotExist$'`
- [ ] **Step 5: Run diff hygiene**
  Run: `git diff --check`
- [ ] **Step 6: Request code review for the completed slice**
- [ ] **Step 7: Commit coherent changes**
  Commit message: `feat: 实现统计与系统日志 P0`
