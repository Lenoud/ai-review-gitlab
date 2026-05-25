# LLM Model Management P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make LLM model configuration real so review workers can later select a default OpenAI-compatible model.

**Architecture:** Follow the existing layered style: `handler -> service -> repository -> model`, with an `llm` adapter for connection checks. Router receives `LLMModelHandler` through dependencies. Database startup includes the `llm_model` table in AutoMigrate.

**Tech Stack:** Gin, GORM, net/http OpenAI-compatible client, sqlite tests, Testify.

---

## Scope

Implement:

- `POST /api/v1/admin/llm-model/create`
- `POST /api/v1/admin/llm-model/update`
- `GET /api/v1/admin/llm-model/get`
- `POST /api/v1/admin/llm-model/delete`
- `GET /api/v1/admin/llm-model/search`
- `GET /api/v1/admin/llm-model/default`
- `POST /api/v1/admin/llm-model/set-default`
- `POST /api/v1/admin/llm-test/connection`

Rules:

- `provider`, `modelCode`, `apiBaseUrl`, and `apiKey` are required for stored models.
- Only one model can be default at a time.
- Deleting the default model is allowed; the next default becomes empty until explicitly set.
- `test-connection` accepts inline config and uses an injectable checker. Tests use fake/check server; production uses OpenAI-compatible `/chat/completions`.

## File Structure

- Create `backend/internal/model/llm_model.go`.
- Create `backend/internal/repository/llm_model_repository.go`.
- Create `backend/internal/service/llm_model_service.go`.
- Create `backend/internal/handler/llm_model_handler.go`.
- Create `backend/internal/llm/checker.go`.
- Modify `backend/internal/router/router.go`.
- Modify `backend/cmd/server/main.go`.
- Modify `backend/internal/database/database.go`.
- Create `backend/migrations/000003_llm_model_schema.sql`.
- Add service/repository/handler/router/llm tests.

---

### Task 1: Service Contract

- [ ] Write failing tests for create, update missing model, search paging, set default, get default, delete empty ids, and test connection.
- [ ] Run `go test ./internal/service -run LLM -count=1` and verify it fails.
- [ ] Implement service types, validation, default switching contract, and checker interface.
- [ ] Run service tests and verify they pass.

### Task 2: Repository And Model

- [ ] Write failing repository tests with sqlite for CRUD, search, default switching, and not found.
- [ ] Run `go test ./internal/repository -run LLM -count=1` and verify it fails.
- [ ] Implement model and GORM repository transaction for `SetDefault`.
- [ ] Run repository tests and verify they pass.

### Task 3: LLM Connection Checker

- [ ] Write failing tests using `httptest.Server` for OpenAI-compatible success and non-2xx failure.
- [ ] Run `go test ./internal/llm -run Checker -count=1` and verify it fails.
- [ ] Implement `checker.go` with a small `/chat/completions` request.
- [ ] Run checker tests and verify they pass.

### Task 4: Handler, Router, Runtime

- [ ] Write failing handler/router tests for create, search, default, set-default, and test-connection.
- [ ] Implement `LLMModelHandler`.
- [ ] Register LLM routes separately from generic stubs.
- [ ] Wire repository/service/handler/checker in `cmd/server`.
- [ ] Add AutoMigrate and SQL migration.
- [ ] Run `go test ./...`.
- [ ] Commit with the project commit-message style, excluding the existing local `config.go` password change.

