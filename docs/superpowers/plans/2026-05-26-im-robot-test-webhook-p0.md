# IM Robot Test Webhook Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move `POST /api/v1/admin/im-robot/test-webhook` off the generic 501 stub and back it with a Java-compatible webhook probe.

**Architecture:** Extend the existing IM robot service with an injectable HTTP client boundary. The service validates platform and webhook URL, sends a small platform-specific text payload, parses DingTalk/WeCom `errcode` or Feishu-style `code` JSON responses, and returns `{success,message}` without persisting data.

**Tech Stack:** Go, Gin, GORM-adjacent service layering, `net/http`, existing response envelopes, testify.

---

### Task 1: Service Contract

**Files:**
- Modify: `backend/internal/service/im_robot_service.go`
- Test: `backend/internal/service/im_robot_service_test.go`

- [x] Write failing tests for successful DingTalk/WeCom-style responses, failed Feishu-style responses, invalid input, malformed JSON, and transport errors using an injected fake sender.
- [x] Run `GOCACHE=/private/tmp/ai-revire-go-cache go test ./internal/service -run IMRobotServiceTestWebhook -count=1` from `backend`; expect compile failures because the method/types do not exist yet.
- [x] Add `IMRobotWebhookSender`, `IMRobotTestWebhookInput`, `IMRobotTestWebhookResult`, default HTTP sender, payload builder, response parser, and `TestWebhook`.
- [x] Run the focused service tests; expect pass.

### Task 2: Handler And Router Wiring

**Files:**
- Modify: `backend/internal/handler/im_robot_handler.go`
- Test: `backend/internal/handler/im_robot_handler_test.go`
- Modify: `backend/internal/router/router.go`
- Test: `backend/internal/router/router_contract_test.go`

- [x] Write failing handler test asserting JSON request binding calls `TestWebhook` and returns a success envelope.
- [x] Update router contract test so `/im-robot/test-webhook` expects a real handler bad-request response instead of 501.
- [x] Run focused handler/router tests; expect compile/status failures.
- [x] Add handler method and register route in `registerIMRobotRoutes`; remove it from generic route definitions.
- [x] Run focused handler/router tests; expect pass.

### Task 3: Verification And Commit

**Files:**
- Modify only files listed above plus this plan.

- [x] Run focused IM robot tests.
- [x] Run the non-network backend package suite.
- [x] Run compile-only all-package test.
- [x] Run `git diff --check`.
- [x] Commit the coherent slice.
