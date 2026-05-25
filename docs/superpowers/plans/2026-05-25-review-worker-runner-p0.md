# Review Worker Runner P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an optional background runner that continuously calls `ReviewWorkerService.ProcessNext` and stops cleanly with the server process.

**Architecture:** Add a small `internal/worker` package for lifecycle concerns only. Keep review execution in `service.ReviewWorkerService`; the runner handles worker ID, polling interval, idle wait, error wait, context cancellation, and goroutine shutdown. Add `worker` config defaults and wire startup in `cmd/server` behind `worker.enabled`.

**Tech Stack:** Go context, sync.WaitGroup, time.Ticker/Timer-style sleeps, existing config/viper, existing service/repository adapters.

---

## Scope

Implement:

- `worker.Runner` with `Start(ctx)` and `Wait()`.
- Config:
  - `worker.enabled` default `false`.
  - `worker.id` default `review-worker-1`.
  - `worker.idle_interval` default `5s`.
  - `worker.error_interval` default `30s`.
- Server wiring when enabled.

Defer:

- Multiple concurrent worker goroutines.
- Running-task timeout recovery.
- Admin API to inspect worker state.
- Structured logging beyond standard `log.Printf`.

## Files

- Create: `backend/internal/worker/runner.go`
- Test: `backend/internal/worker/runner_test.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/config.example.yaml`

---

### Task 1: Runner

- [x] Write failing tests for runner loop:
  - processes immediately on start.
  - waits on idle result before polling again.
  - waits on error before polling again.
  - exits when context is canceled.
- [x] Run `go test ./internal/worker -count=1` and verify it fails because package/code does not exist.
- [x] Implement runner with injectable sleeper for fast tests.
- [x] Run runner tests and verify they pass.

### Task 2: Config

- [x] Add failing config assertions for worker defaults and env overrides.
- [x] Run `go test ./internal/config -count=1` and verify it fails.
- [x] Implement `WorkerConfig` defaults.
- [x] Update `config.example.yaml`.
- [x] Run config tests and verify they pass.

### Task 3: Server Wiring

- [x] Wire `ReviewWorkerService` and `worker.Runner` in `cmd/server/main.go` only when `cfg.Worker.Enabled` is true.
- [x] Use existing repositories/adapters: review task service, project repository, LLM model repository, GitLab service adapter, OpenAI-compatible chat client.
- [x] Run full `go test ./...`.
- [x] Commit with project `feat:` style.
