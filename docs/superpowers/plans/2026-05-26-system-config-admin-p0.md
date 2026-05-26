# System Config Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move admin system info/config/base-url endpoints off generic 501 stubs and back them with persisted settings.

**Architecture:** Follow the existing Go backend layering: `handler -> service -> repository -> model`. Store settings in a `settings` key/value table, using key `SYSTEM` with a JSON value compatible with the Java reference (`version`, `siteName`, `siteNotice`, `baseUrl`). Keep cloud license/version/value-added endpoints out of scope.

**Tech Stack:** Go, Gin, GORM, MySQL migrations, testify.

---

## File Structure

- Create `backend/internal/model/setting.go`: GORM model for the `settings` table.
- Create `backend/internal/repository/setting_repository.go` and tests: read/write keyed setting values.
- Create `backend/internal/service/system_service.go` and tests: defaults, persisted config parsing, base URL update validation.
- Create `backend/internal/handler/system_handler.go` and tests: admin/open HTTP handlers.
- Modify `backend/internal/router/router.go` and `router_contract_test.go`: wire system routes to real handler.
- Modify `backend/cmd/server/main.go`: construct repository/service/handler dependencies.
- Modify `backend/internal/database/database.go`: add AutoMigrate model.
- Add `backend/migrations/000013_settings_schema.sql`.

## Tasks

### Task 1: Repository and Model

- [ ] Add failing repository tests for missing value, create value, and update existing value.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run Setting -count=1` from `backend`; expect failure because repository/model do not exist.
- [ ] Add `model.Setting`, `SettingRepository`, `GetSettingValue`, and `SetSettingValue`.
- [ ] Add migration and AutoMigrate entry.
- [ ] Re-run repository tests; expect pass.

### Task 2: Service

- [ ] Add failing service tests for default config, JSON-backed config, invalid JSON failure, valid base URL update, blank base URL rejection, and non-HTTP URL rejection.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run System -count=1`; expect failure.
- [ ] Implement `SystemService` with default `version=1.0.0`, `siteName=AI Code Review`, `siteNotice=""`, and optional `baseUrl`.
- [ ] Validate `baseUrl` as absolute `http` or `https` URL with host and trim trailing slash.
- [ ] Re-run service tests; expect pass.

### Task 3: Handler and Router

- [ ] Add failing handler tests for open info, admin config, admin info, and update base URL.
- [ ] Update router contract tests so `/admin/system/info`, `/admin/system/config`, and `/admin/system/config/base-url` no longer expect 501.
- [ ] Run focused handler/router tests; expect failure.
- [ ] Implement `SystemHandler` and wire dependencies in router/main.
- [ ] Re-run focused tests; expect pass.

### Task 4: Verification and Commit

- [ ] Run focused system tests.
- [ ] Run non-network package suite:
  `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/config ./internal/database ./internal/handler ./internal/middleware ./internal/pkg/response ./internal/repository ./internal/router ./internal/service ./internal/worker -count=1`
- [ ] Run compile-only all-package check:
  `GOCACHE=/private/tmp/ai-revire-go-cache go test ./... -run '^TestDoesNotExist$'`
- [ ] Run `git diff --check` and staged diff check.
- [ ] Commit as `feat: 实现系统配置管理 P0`.
