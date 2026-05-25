# Review Worker P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Execute one queued review task end-to-end: claim, load project/model, fetch GitLab diff, call LLM, and update task status.

**Architecture:** Keep the existing layered style. Add `service.ReviewWorkerService` as the orchestration layer, reuse `ReviewTaskService` for queue state transitions, reuse `ProjectRepository` and `LLMModelRepository` for configuration, and depend on small GitLab/LLM interfaces for external calls. Add an OpenAI-compatible chat client under `internal/llm`.

**Tech Stack:** Go, Gin-adjacent service layer, net/http, GORM repositories, httptest, Testify.

---

## Scope

Implement:

- Single-task worker method: `ProcessNext(ctx, workerID)`.
- Push task diff loading via GitLab commit diff.
- Merge request task diff loading via GitLab MR changes.
- Default LLM model lookup and OpenAI-compatible chat completion call.
- Success and retry/failure transitions through existing `ReviewTaskService`.

Defer:

- Background goroutine startup loop.
- Review log tables and APIs.
- GitLab comment posting.
- HTML report generation.
- IM notification.
- Attempt row success/failure updates beyond existing task state.

## Files

- Create: `backend/internal/service/review_worker_service.go`
- Test: `backend/internal/service/review_worker_service_test.go`
- Create: `backend/internal/llm/client.go`
- Test: `backend/internal/llm/client_test.go`
- Modify: `backend/cmd/server/main.go` only if a compile-time construction hook is needed.

---

### Task 1: Worker Service Contract

- [x] Write failing tests for `ReviewWorkerService.ProcessNext`:
  - returns `processed=false` when no pending task exists.
  - for push tasks, claims task, starts attempt, fetches commit diff, calls LLM, marks succeeded.
  - when LLM fails, marks failed through existing retry logic.
- [x] Run `go test ./internal/service -run ReviewWorker -count=1` and verify it fails because the service does not exist.
- [x] Implement minimal worker service and GitLab webhook payload parsing local to worker.
- [x] Run service tests and verify they pass.

### Task 2: OpenAI-Compatible Chat Client

- [x] Write failing `httptest.Server` tests for chat completion request/response and non-2xx error.
- [x] Run `go test ./internal/llm -run Chat -count=1` and verify it fails.
- [x] Implement `OpenAICompatibleClient.Chat`.
- [x] Run LLM tests and verify they pass.

### Task 3: Verification And Commit

- [x] Run targeted tests: `go test ./internal/service ./internal/llm -count=1`.
- [x] Run full tests: `go test ./...`.
- [x] Stage only worker P0 files.
- [x] Commit with project `feat:` style.
