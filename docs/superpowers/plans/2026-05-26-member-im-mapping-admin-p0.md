# Member IM Mapping Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move member IM mapping create/update/get/delete/search admin routes off generic 501 stubs.

**Architecture:** Follow the existing backend layering used by IM robot management: Gin handler parses HTTP input, service normalizes and validates fields, repository persists `member_im_mapping` rows through GORM, and router wires implemented routes with existing RBAC permissions. This slice only stores Git member to IM user IDs; notification delivery and commit-summary behavior remain separate.

**Tech Stack:** Go, Gin, GORM, MySQL migrations, sqlite-backed repository tests, testify.

---

### File Structure

- Create: `backend/internal/model/member_im_mapping.go` for the GORM model.
- Create: `backend/internal/service/member_im_mapping_service.go` and `_test.go` for validation, pagination, and service behavior.
- Create: `backend/internal/repository/member_im_mapping_repository.go` and `_test.go` for persistence, uniqueness checks, and search.
- Create: `backend/internal/handler/member_im_mapping_handler.go` and `_test.go` for HTTP request/response/error mapping.
- Create: `backend/migrations/000011_member_im_mapping_schema.sql` for database schema.
- Modify: `backend/internal/database/database.go` to add AutoMigrate registration.
- Modify: `backend/internal/router/router.go` and `router_contract_test.go` to route five member mapping endpoints to the real handler.
- Modify: `backend/cmd/server/main.go` to construct repository/service/handler dependencies.

### Task 1: Service Contract

- [x] Write failing service tests for create normalization, duplicate protection, explicit validation errors, delete ID cleanup, and search pagination defaults.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run MemberIMMapping -count=1` from `backend`; expect compile failure because service types do not exist.
- [x] Implement `MemberIMMappingService`, DTOs, errors, and repository interface.
- [x] Re-run the service test command; expect PASS.

### Task 2: Repository

- [x] Write failing repository tests for create/update/get/search/delete, duplicate lookup, and not-found mapping.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run MemberIMMapping -count=1`; expect compile failure because model/repository do not exist.
- [x] Implement model, repository, and migration; register model in AutoMigrate.
- [x] Re-run the repository test command; expect PASS.

### Task 3: Handler And Router

- [x] Write failing handler tests for create/update/get/search/delete and error mapping.
- [x] Update router contract expectations so member mapping routes no longer expect generic 501.
- [x] Run focused handler/router tests; expect compile failure or status mismatch until wiring exists.
- [x] Implement handler, router registration, and server DI wiring.
- [x] Re-run focused handler/router tests; expect PASS.

### Task 4: Verification And Review

- [x] Run focused package tests:
  `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service ./internal/repository ./internal/handler ./internal/router -run MemberIMMapping -count=1`
- [x] Run non-network touched package suite:
  `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/config ./internal/database ./internal/handler ./internal/middleware ./internal/pkg/response ./internal/repository ./internal/router ./internal/service ./internal/worker -count=1`
- [x] Run compile-only all-package check:
  `GOCACHE=/private/tmp/ai-revire-go-cache go test ./... -run '^TestDoesNotExist$'`
- [x] Run `git diff --check`.
- [x] Request code review and fix Critical/Important findings before commit. Review agent timed out and was closed; local focused review found no blocking issues.
- [x] Commit as a coherent P0 slice.
