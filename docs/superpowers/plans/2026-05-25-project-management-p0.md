# Project Management P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the core project management endpoints real so later Webhook, GitLab, LLM, and notification flows have a stable project configuration source.

**Architecture:** Keep the existing layered shape: `handler -> service -> repository -> model`. The router receives a `ProjectHandler` dependency and only the project-management routes are moved off the generic `501` stub. GitLab remote/group search and prompt rendering remain stubbed in this slice.

**Tech Stack:** Gin, GORM, MySQL/sqlite tests, Testify.

---

## Scope

Implement:

- `POST /api/v1/admin/project/create`
- `POST /api/v1/admin/project/batch-create`
- `POST /api/v1/admin/project/update`
- `GET /api/v1/admin/project/get`
- `POST /api/v1/admin/project/delete`
- `GET /api/v1/admin/project/search`
- `POST /api/v1/admin/project/web-urls/exists`

Defer:

- `project/gitlab/remote-search`
- `project/gitlab/group-search`
- project review prompt endpoints

## File Structure

- Create `backend/internal/model/project.go`: GORM project model.
- Create `backend/internal/repository/project_repository.go`: project persistence.
- Create `backend/internal/service/project_service.go`: validation and business rules.
- Create `backend/internal/handler/project_handler.go`: Gin request binding and response writing.
- Modify `backend/internal/router/router.go`: inject and register project handler methods.
- Modify `backend/internal/database/database.go`: include project model in AutoMigrate.
- Create `backend/migrations/000002_project_schema.sql`: SQL schema.
- Add focused tests for service, repository, handler, and router contract.

---

### Task 1: Service Contract

- [ ] Write failing tests for create, duplicate URL rejection, update missing project, delete empty ids, search paging, and URL exists.
- [ ] Run `go test ./internal/service -run Project -count=1` and verify it fails.
- [ ] Implement service DTOs, repository interface, validation, and pagination defaults.
- [ ] Run the service tests and verify they pass.

### Task 2: Repository And Model

- [ ] Write failing repository tests with sqlite in-memory DB for CRUD, search, batch create, and URL existence.
- [ ] Run `go test ./internal/repository -run Project -count=1` and verify it fails.
- [ ] Implement model and GORM repository.
- [ ] Run repository tests and verify they pass.

### Task 3: Handler And Router

- [ ] Write failing handler/router tests for create/get/search/delete/web-url-exists.
- [ ] Run targeted handler/router tests and verify they fail.
- [ ] Implement project handler and route injection.
- [ ] Run targeted tests and verify they pass.

### Task 4: Runtime Wiring

- [ ] Add project model to AutoMigrate.
- [ ] Add `000002_project_schema.sql`.
- [ ] Wire project repository/service/handler in `cmd/server`.
- [ ] Run `go test ./...`.
- [ ] Commit with the project commit-message style.

