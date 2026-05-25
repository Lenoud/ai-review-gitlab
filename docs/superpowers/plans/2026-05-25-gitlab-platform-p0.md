# GitLab Platform Adapter P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first real GitLab API adapter and wire remote project/group search endpoints.

**Architecture:** Add `internal/platform/gitlab` as the HTTP adapter. Keep API orchestration in `service.GitLabService`, and keep request/response binding in `handler.ProjectGitLabHandler`. Router injects this handler and removes `remote-search` / `group-search` from generic stubs.

**Tech Stack:** Gin, net/http, httptest, Testify.

---

## Scope

Implement:

- `POST /api/v1/admin/project/gitlab/remote-search`
- `POST /api/v1/admin/project/gitlab/group-search`
- GitLab client methods:
  - Search projects: `GET /api/v4/projects?search=...&page=...&per_page=...&simple=true`
  - Search groups: `GET /api/v4/groups?search=...&page=...&per_page=...`
  - Get MR changes: `GET /api/v4/projects/:id/merge_requests/:iid/changes`
  - Get commit diff: `GET /api/v4/projects/:id/repository/commits/:sha/diff`

Defer:

- Worker integration
- Posting comments back to GitLab
- Pagination metadata parsing beyond returning the current page

## File Structure

- Create `backend/internal/platform/gitlab/client.go`.
- Create `backend/internal/service/gitlab_service.go`.
- Create `backend/internal/handler/project_gitlab_handler.go`.
- Modify `backend/internal/router/router.go`.
- Modify `backend/cmd/server/main.go`.
- Add tests for platform client, service validation, handler, and router contract.

---

### Task 1: GitLab Client

- [x] Write failing `httptest.Server` tests for project search, group search, MR changes, commit diff, and non-2xx error.
- [x] Run `go test ./internal/platform/gitlab -count=1` and verify it fails.
- [x] Implement client with private-token header, URL joining, query params, and JSON decoding.
- [x] Run client tests and verify they pass.

### Task 2: Service

- [x] Write failing tests for input validation and delegation to client.
- [x] Run `go test ./internal/service -run GitLab -count=1` and verify it fails.
- [x] Implement `GitLabService`.
- [x] Run service tests and verify they pass.

### Task 3: Handler And Router

- [x] Write failing handler tests for remote-search and group-search.
- [x] Update router contract expected statuses for both routes.
- [x] Implement handler and route registration.
- [x] Wire handler in `cmd/server`.
- [x] Run targeted tests and full `go test ./...`.
- [x] Commit with the project commit-message style.
