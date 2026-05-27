# IM Validation Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Return validation errors for IM robot and member IM mapping string fields that exceed database column limits.

**Architecture:** Keep validation in the existing service normalization functions so handlers consistently map bad input to HTTP 400 before any repository/database write. Reuse model/migration limits already present in the codebase and avoid changing persistence schemas.

**Tech Stack:** Go, Gin handlers, GORM repositories, testify service tests.

---

### Task 1: Member IM Mapping Length Validation

**Files:**
- Modify: `backend/internal/service/member_im_mapping_service_test.go`
- Modify: `backend/internal/service/member_im_mapping_service.go`

- [x] Write a failing service test that overlong `gitUsername`, `imUserId`, and `displayName` return `ErrInvalidMemberIMMappingInput`.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run 'TestMemberIMMappingServiceRejectsOverlongInput' -count=1` from `backend` and confirm it fails.
- [x] Add service constants and length checks matching `member_im_mapping` limits: git username 128, platform 32, IM user ID 256, display name 128.
- [x] Re-run the focused service test and confirm it passes.

### Task 2: IM Robot Length Validation

**Files:**
- Modify: `backend/internal/service/im_robot_service_test.go`
- Modify: `backend/internal/service/im_robot_service.go`

- [x] Write a failing service test that overlong `name`, `webhookUrl`, and `secret` return `ErrInvalidIMRobotInput`.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run 'TestIMRobotServiceRejectsOverlongInput' -count=1` from `backend` and confirm it fails.
- [x] Add service constants and length checks matching `im_robot` limits: platform 32, name 128, webhook URL 1024, secret 512.
- [x] Re-run the focused service test and confirm it passes.

### Task 3: Verify And Commit

**Files:**
- Verify touched service tests and non-network package suite.

- [x] Run focused IM service tests after formatting.
- [x] Run non-network package suite after formatting.
- [x] Run compile-only all-package check after formatting.
- [x] Run `git diff --check` after formatting.
- [x] Commit the coherent validation hardening change.
