# IM Robot Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move IM robot create/update/get/delete/search/list-enabled admin routes off generic 501 stubs.

**Architecture:** Follow the existing backend layering: Gin handler parses requests, service normalizes/validates inputs, repository persists `im_robot` rows through GORM, and router wires implemented routes with existing RBAC permissions. Keep `/im-robot/test-webhook` stubbed because live webhook delivery is separate networked behavior.

**Tech Stack:** Go, Gin, GORM, SQLite tests, existing response envelope helpers.

---

### File Map

- Create: `backend/internal/model/im_robot.go` for the `im_robot` GORM model.
- Create: `backend/internal/service/im_robot_service.go` and `_test.go` for validation, defaults, search query normalization, and delete ID cleanup.
- Create: `backend/internal/repository/im_robot_repository.go` and `_test.go` for CRUD/search/list-enabled persistence.
- Create: `backend/internal/handler/im_robot_handler.go` and `_test.go` for request parsing and response/error mapping.
- Create: `backend/migrations/000010_im_robot_schema.sql` for the database schema.
- Modify: `backend/internal/database/database.go` to include `model.IMRobot` in AutoMigrate.
- Modify: `backend/cmd/server/main.go` to construct and inject the handler.
- Modify: `backend/internal/router/router.go` and `router_contract_test.go` to route six IM robot endpoints to the real handler while keeping `/test-webhook` stubbed.

### Task 1: Service Contract

- [ ] Write failing service tests for create normalization, invalid input, update/get/delete ID validation, delete ID dedupe, search query normalization, and enabled default behavior.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run IMRobot -count=1` from `backend`; expect compile failure because the service does not exist.
- [ ] Implement `IMRobotService`, types, repository interface, and errors.
- [ ] Re-run the service tests; expect pass.

### Task 2: Repository Persistence

- [ ] Write failing repository tests for create/update/find, not found mapping, search paging/filtering, list-enabled filtering, delete, and explicit `enabled=false` persistence.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run IMRobot -count=1`; expect compile failure because repository/model do not exist.
- [ ] Add `model.IMRobot` and `IMRobotRepository` using GORM.
- [ ] Re-run repository tests; expect pass.

### Task 3: Handler and Router

- [ ] Write failing handler tests for create/get/search/list-enabled/delete and error mapping.
- [ ] Update router contract expectations so six IM robot routes no longer expect generic 501 with a dev token.
- [ ] Run focused handler/router tests; expect compile failure or status mismatch.
- [ ] Add `IMRobotHandler`, register routes, and wire dependencies in `main.go`.
- [ ] Re-run focused handler/router tests; expect pass.

### Task 4: Migration and Verification

- [ ] Add `backend/migrations/000010_im_robot_schema.sql` and AutoMigrate registration.
- [ ] Run focused IM robot tests.
- [ ] Run non-network package suite: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/config ./internal/database ./internal/handler ./internal/middleware ./internal/pkg/response ./internal/repository ./internal/router ./internal/service ./internal/worker -count=1`.
- [ ] Run compile-only all-package check: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./... -run '^TestDoesNotExist$'`.
- [ ] Run `git diff --check`.
- [ ] Commit as `feat: 实现 IM 机器人管理 P0`.
