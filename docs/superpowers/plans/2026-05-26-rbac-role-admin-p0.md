# RBAC Role Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move admin role create/update/get/delete endpoints off generic 501 stubs.

**Architecture:** Extend the existing RBAC read slice instead of adding a parallel repository. The handler parses role payloads, the service validates and normalizes role fields and permission IDs, and `UserRepository` performs role and role-permission writes inside transactions. User management remains out of scope for this increment.

**Tech Stack:** Go, Gin, GORM, SQLite-backed unit tests.

---

### Task 1: Service Contract And Validation

**Files:**
- Modify: `backend/internal/service/rbac_service.go`
- Modify: `backend/internal/service/auth_service_test.go`

- [x] Add failing service tests for create/update/get/delete validation.
- [x] Add `RoleDetail`, `RoleInput`, and repository interface methods.
- [x] Validate non-empty `code`/`name`, bounded lengths matching schema, positive role IDs, and deduped positive permission IDs.

### Task 2: Repository Persistence

**Files:**
- Modify: `backend/internal/repository/user_repository.go`
- Modify: `backend/internal/repository/user_repository_test.go`

- [x] Add failing repository tests for create with permissions, update replacing permissions, get detail, duplicate code conflict, and delete cleanup.
- [x] Implement transactional role creation/update and role-permission replacement.
- [x] Map missing role and duplicate code to service errors.
- [x] Prevent deleting roles assigned to users.

### Task 3: Handler And Router

**Files:**
- Modify: `backend/internal/handler/rbac_handler.go`
- Modify: `backend/internal/handler/rbac_handler_test.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_contract_test.go`

- [x] Add failing handler tests for role create/update/get/delete.
- [x] Wire `POST /role/create`, `POST /role/update`, `GET /role/get`, and `POST /role/delete`.
- [x] Map invalid input to 400, not found to 404, duplicate/in-use to 409, and repository failures to 500.

### Task 4: Verification

**Commands:**
- [ ] `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service ./internal/repository ./internal/handler ./internal/router -run 'RBAC|UserRepository|AdminRoutes' -count=1`
- [ ] `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/config ./internal/database ./internal/handler ./internal/middleware ./internal/pkg/response ./internal/repository ./internal/router ./internal/service ./internal/worker -count=1`
- [ ] `GOCACHE=/private/tmp/ai-revire-go-cache go test ./... -run '^TestDoesNotExist$'`
- [ ] `git diff --check`
