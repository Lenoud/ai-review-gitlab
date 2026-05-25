# Review Task Queue P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist GitLab webhook events as durable review tasks and provide the first worker-facing task state transitions.

**Architecture:** Keep the existing layered style. `handler.WebhookHandler` parses webhook payloads and delegates to `service.ReviewTaskService`; service builds dedupe keys, finds projects by GitLab web URL, and creates or returns existing tasks; repository owns GORM persistence for `review_task` and `review_task_attempt`.

**Tech Stack:** Gin, GORM, MySQL/sqlite tests, JSON payload storage, Testify.

---

## Scope

Implement:

- `POST /review/webhook` accepts GitLab `Push Hook` and `Merge Request Hook`.
- Webhook returns `{ taskId, duplicate, status }` and never kicks off real review work in this slice.
- Deduplication:
  - Push: `gitlab:push:{projectId}:{ref}:{afterSha}`
  - MR: `gitlab:mr:{projectId}:{mergeRequestIid}:{lastCommitSha}:{action}`
- `review_task` and `review_task_attempt` models, repository, AutoMigrate, SQL migration.
- Worker-facing repository/service primitives:
  - claim pending task
  - start attempt
  - mark succeeded
  - mark failed with retry/backoff

Defer:

- Real GitLab diff fetching
- LLM review execution
- background worker goroutine startup
- review log persistence

## File Structure

- Create `backend/internal/model/review_task.go`.
- Create `backend/internal/repository/review_task_repository.go`.
- Create `backend/internal/service/review_task_service.go`.
- Create `backend/internal/handler/webhook_handler.go`.
- Modify `backend/internal/repository/project_repository.go`: add lookup by web URL.
- Modify `backend/internal/service/project_service.go`: extend repository interface.
- Modify `backend/internal/router/router.go`: inject webhook handler.
- Modify `backend/cmd/server/main.go`: wire review task service/handler.
- Modify `backend/internal/database/database.go`: include review task models.
- Create `backend/migrations/000004_review_task_schema.sql`.

---

### Task 1: Service Contract

- [ ] Write failing tests for push webhook dedupe, MR webhook dedupe, unknown project rejection, and failed task retry scheduling.
- [ ] Run `go test ./internal/service -run 'ReviewTask|Webhook' -count=1` and verify it fails.
- [ ] Implement service DTOs, dedupe key generation, and task status transitions.
- [ ] Run service tests and verify they pass.

### Task 2: Repository And Model

- [ ] Write failing repository tests with sqlite for create-or-get duplicate, claim pending task, attempt creation, succeeded, failed retry, and failed final state.
- [ ] Run `go test ./internal/repository -run ReviewTask -count=1` and verify it fails.
- [ ] Implement models and GORM repository.
- [ ] Run repository tests and verify they pass.

### Task 3: Webhook Handler

- [ ] Write failing handler tests for push and merge request payloads.
- [ ] Run `go test ./internal/handler -run Webhook -count=1` and verify it fails.
- [ ] Implement `WebhookHandler`.
- [ ] Run handler tests and verify they pass.

### Task 4: Router And Runtime

- [ ] Move `/review/webhook` from generic `501` to the injected `WebhookHandler`.
- [ ] Wire repository/service/handler in `cmd/server`.
- [ ] Add AutoMigrate models and SQL migration.
- [ ] Run `go test ./...`.
- [ ] Commit with the project commit-message style.

