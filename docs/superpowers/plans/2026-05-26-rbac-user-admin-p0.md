# RBAC User Admin P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move admin user create/update/get/search/role-options endpoints off generic 501 stubs.

**Architecture:** Extend the existing RBAC slice instead of creating a separate user-admin module. `RBACHandler` handles request parsing/errors, `RBACService` validates and hashes passwords, and `UserRepository` owns `sys_user` plus `sys_user_role` persistence.

**Tech Stack:** Go, Gin, GORM, SQLite repository tests, existing response envelope.

---

### Task 1: Service Contract And Validation

**Files:**
- Modify: `backend/internal/service/rbac_service.go`
- Modify: `backend/internal/service/auth_service_test.go`

- [ ] Add service tests for create/update/search normalization: username/password/nickname/remark length limits, role ID dedupe, password hash generation, optional update password, page/size defaults.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run 'RBACService.*User|AuthService' -count=1` from `backend`; expect compile failures for missing user types/methods.
- [ ] Add `AdminUser`, `AdminUserInput`, `AdminUserSearchQuery`, `AdminUserPage`, user errors, and RBAC repository interface methods.
- [ ] Implement `CreateUser`, `UpdateUser`, `GetUser`, `SearchUsers`, and `ListRoleOptions` in `RBACService`.
- [ ] Run the focused service test until it passes.

### Task 2: Repository Persistence

**Files:**
- Modify: `backend/internal/repository/user_repository.go`
- Modify: `backend/internal/repository/user_repository_test.go`

- [ ] Add repository tests for creating users with roles, updating users while preserving password when blank, searching by username/nickname, duplicate username, unknown role IDs, and role options ordering.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/repository -run 'UserRepository.*User|UserRepository.*RoleOptions' -count=1` from `backend`; expect compile failures for missing repository methods.
- [ ] Implement transactional user create/update, role validation, role binding replacement, user search with role summaries, and username uniqueness mapping.
- [ ] Run the focused repository tests until they pass.

### Task 3: Handler And Router Wiring

**Files:**
- Modify: `backend/internal/handler/rbac_handler.go`
- Modify: `backend/internal/handler/rbac_handler_test.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_contract_test.go`

- [ ] Add handler tests for user create/update/get/search/role-options and error mappings.
- [ ] Update router contract expectations so `/api/v1/admin/user/*` routes no longer expect generic 501.
- [ ] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/handler ./internal/router -run 'RBAC|AdminRoutes' -count=1` from `backend`; expect compile/status failures.
- [ ] Add RBAC handler methods and register user routes with `rbac:read`/`rbac:write`.
- [ ] Run focused handler/router tests until they pass.

### Task 4: Verification And Commit

**Files:**
- All touched files.

- [ ] Run `gofmt` on modified Go files.
- [ ] Run focused RBAC tests: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service ./internal/repository ./internal/handler ./internal/router -run 'RBAC|UserRepository|AdminRoutes' -count=1`.
- [ ] Run non-network package suite: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/config ./internal/database ./internal/handler ./internal/middleware ./internal/pkg/response ./internal/repository ./internal/router ./internal/service ./internal/worker -count=1`.
- [ ] Run compile-only all-package check: `GOCACHE=/private/tmp/ai-revire-go-cache go test ./... -run '^TestDoesNotExist$'`.
- [ ] Run `git diff --check`.
- [ ] Request code review for the completed slice and fix Critical/Important findings.
- [ ] Commit the coherent slice, excluding unrelated `.gitignore` and `java-src/` changes.
