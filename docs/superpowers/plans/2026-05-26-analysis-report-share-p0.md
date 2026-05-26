# Analysis Report Share P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Serve shared project analysis execution reports from `GET /api/v1/open/analysis-report` with token validation.

**Architecture:** Mirror the existing public code review report path: handler parses `logId` and `token`, service validates token and expiry, repository loads the execution log, and router wires the public route to the injected handler. Add a minimal `project_analysis_plan_execution_log` model/repository surface needed for public report reads, while keeping admin analysis-plan management stubbed.

**Tech Stack:** Go, Gin, GORM, sqlite tests, MySQL migration SQL.

---

### Task 1: Service Contract

**Files:**
- Modify: `backend/internal/service/open_report_service.go`
- Modify: `backend/internal/service/open_report_service_test.go`

- [ ] **Step 1: Write failing service test**

Add a test that creates a fake analysis execution log with `ShareToken`, future `ShareTokenExpiresAt`, `ResultContent` containing a `<script>` tag, and verifies `OpenReportService.AnalysisReport` returns escaped HTML containing the result content.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run OpenReport -count=1`
Expected: FAIL because `AnalysisReport` and analysis-log types are missing.

- [ ] **Step 3: Implement minimal service API**

Add `AnalysisReportInput`, `ProjectAnalysisPlanExecutionLog`, `FindAnalysisExecutionByID` to the report repository interface, `AnalysisReport`, and a small HTML builder using the same token validation and escaping policy as code review reports.

- [ ] **Step 4: Run service tests**

Run: `go test ./internal/service -run OpenReport -count=1`
Expected: PASS.

### Task 2: Repository And Migration

**Files:**
- Create: `backend/internal/model/project_analysis_plan.go`
- Create: `backend/internal/repository/analysis_execution_log_repository_test.go`
- Modify: `backend/internal/repository/review_log_repository.go`
- Modify: `backend/internal/database/database.go`
- Create: `backend/migrations/000007_project_analysis_plan_schema.sql`

- [ ] **Step 1: Write failing repository test**

Add sqlite repository test for `FindAnalysisExecutionByID`, including not-found mapping to `service.ErrReviewLogNotFound`.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/repository -run AnalysisExecution -count=1`
Expected: FAIL because the model/repository method is missing.

- [ ] **Step 3: Implement model and repository method**

Add `ProjectAnalysisPlanExecutionLog` GORM model, conversion helper, repository lookup method, and include the model in `database.AutoMigrate`.

- [ ] **Step 4: Add SQL migration**

Create the P0 schema for `project_analysis_plan` and `project_analysis_plan_execution_log` matching the design fields needed now plus planned admin routes.

- [ ] **Step 5: Run repository/database tests**

Run: `go test ./internal/repository ./internal/database -count=1`
Expected: PASS.

### Task 3: Handler And Router

**Files:**
- Modify: `backend/internal/handler/open_report_handler.go`
- Modify: `backend/internal/handler/open_report_handler_test.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_contract_test.go`

- [ ] **Step 1: Write failing handler/router tests**

Add a handler test for `GET /analysis-report?logId=...&token=...` and update router contract expectations so the public route is no longer a generic 501 stub.

- [ ] **Step 2: Run tests to verify failure**

Run: `go test ./internal/handler ./internal/router -run 'OpenReport|PublicRoutes' -count=1`
Expected: FAIL because the handler and route method are missing.

- [ ] **Step 3: Implement handler and route wiring**

Add `OpenReportService.AnalysisReport` to the handler interface, implement `AnalysisReport`, and wire `/api/v1/open/analysis-report` to `deps.OpenReportHandler.AnalysisReport`.

- [ ] **Step 4: Run targeted tests**

Run: `go test ./internal/handler ./internal/router -run 'OpenReport|PublicRoutes' -count=1`
Expected: PASS.

### Task 4: Verification And Commit

**Files:**
- All files above.

- [ ] **Step 1: Format**

Run: `gofmt -w backend/internal/service/open_report_service.go backend/internal/service/open_report_service_test.go backend/internal/model/project_analysis_plan.go backend/internal/repository/review_log_repository.go backend/internal/repository/analysis_execution_log_repository_test.go backend/internal/database/database.go backend/internal/handler/open_report_handler.go backend/internal/handler/open_report_handler_test.go backend/internal/router/router.go backend/internal/router/router_contract_test.go`

- [ ] **Step 2: Full verification**

Run: `go test ./...`
Expected: PASS.

- [ ] **Step 3: Commit coherent slice**

Run: `git add docs/superpowers/plans/2026-05-26-analysis-report-share-p0.md backend/internal/service/open_report_service.go backend/internal/service/open_report_service_test.go backend/internal/model/project_analysis_plan.go backend/internal/repository/review_log_repository.go backend/internal/repository/analysis_execution_log_repository_test.go backend/internal/database/database.go backend/internal/handler/open_report_handler.go backend/internal/handler/open_report_handler_test.go backend/internal/router/router.go backend/internal/router/router_contract_test.go backend/migrations/000007_project_analysis_plan_schema.sql && git commit -m "feat: 实现分析报告分享 P0"`
