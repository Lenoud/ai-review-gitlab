# P0 API Contract Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a runnable Go/Gin backend skeleton that registers the V1 API surface, returns consistent envelopes, enforces basic route grouping, and marks unfinished business endpoints as explicit `501 Not Implemented`.

**Architecture:** This first slice builds only the contract layer: app bootstrap, config defaults, router, middleware, response helpers, DTO placeholders, and handler stubs. Business logic, database access, GitLab, LLM, notification, migrations, and workers are intentionally deferred; handlers must not fake success for unfinished behavior.

**Tech Stack:** Go 1.22+, Gin, Viper, Zap, JWT library placeholder dependency, bcrypt dependency placeholder, Testify.

---

## File Structure

- Create `backend/go.mod`: module definition and initial dependencies.
- Create `backend/cmd/server/main.go`: load config, build router, start HTTP server.
- Create `backend/internal/config/config.go`: minimal config struct and default loading.
- Create `backend/internal/pkg/response/response.go`: uniform success/error JSON helpers.
- Create `backend/internal/pkg/errors/errors.go`: app error type and common codes.
- Create `backend/internal/middleware/recovery.go`: Gin recovery middleware wrapper.
- Create `backend/internal/middleware/request_logger.go`: request logging middleware.
- Create `backend/internal/middleware/auth.go`: temporary JWT/RBAC middleware that distinguishes public/admin routes without real token issuance.
- Create `backend/internal/handler/stub.go`: `NotImplemented` handler and simple public/system handlers.
- Create `backend/internal/router/router.go`: route registration for every endpoint in the spec.
- Create `backend/internal/dto/README.md`: document that DTOs will be filled as each endpoint becomes real.
- Create `backend/internal/router/router_contract_test.go`: route contract tests.
- Create `backend/internal/pkg/response/response_test.go`: response helper tests.
- Create `backend/Makefile`: common commands.

The skeleton should compile, run, and expose all listed routes. Unimplemented endpoints return:

```json
{
  "code": 50100,
  "message": "接口暂未实现"
}
```

Public health/system endpoints may return real minimal data. Authenticated admin endpoints should go through auth middleware before reaching `501`.

---

### Task 1: Initialize Go Module And Tooling

**Files:**
- Create: `backend/go.mod`
- Create: `backend/Makefile`

- [ ] **Step 1: Create failing module smoke test command**

Run:

```bash
go test ./...
```

Expected: FAIL because no Go module exists.

- [ ] **Step 2: Add module and dependencies**

Create `backend/go.mod`:

```go
module github.com/Lenoud/ai-review-gitlab/backend

go 1.22

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.9.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.26.0
)
```

Create `backend/Makefile`:

```makefile
.PHONY: test run fmt

test:
	go test ./...

run:
	go run ./cmd/server

fmt:
	gofmt -w ./cmd ./internal
```

- [ ] **Step 3: Verify module resolves**

Run:

```bash
go mod tidy
go test ./...
```

Expected: PASS or `go test` reports no packages yet. `backend/go.sum` is created.

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum Makefile
git commit -m "chore: initialize go module"
```

---

### Task 2: Add Response And Error Helpers

**Files:**
- Create: `backend/internal/pkg/errors/errors.go`
- Create: `backend/internal/pkg/response/response.go`
- Test: `backend/internal/pkg/response/response_test.go`

- [ ] **Step 1: Write failing response tests**

Create `backend/internal/pkg/response/response_test.go`:

```go
package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSuccessWritesEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Success(c, gin.H{"ok": true})

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.EqualValues(t, 0, body["code"])
	require.Equal(t, "success", body["message"])
	require.Equal(t, true, body["data"].(map[string]any)["ok"])
}

func TestNotImplementedWrites501Envelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.NotImplemented(c)

	require.Equal(t, http.StatusNotImplemented, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.EqualValues(t, 50100, body["code"])
	require.Equal(t, "接口暂未实现", body["message"])
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/pkg/response -run 'TestSuccessWritesEnvelope|TestNotImplementedWrites501Envelope' -v
```

Expected: FAIL because packages do not exist.

- [ ] **Step 3: Implement helpers**

Create `backend/internal/pkg/errors/errors.go`:

```go
package errors

const (
	CodeSuccess        = 0
	CodeBadRequest    = 40000
	CodeUnauthorized  = 40100
	CodeForbidden     = 40300
	CodeNotFound      = 40400
	CodeInternal      = 50000
	CodeNotImplemented = 50100
)

type AppError struct {
	HTTPStatus int
	Code       int
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}
```

Create `backend/internal/pkg/response/response.go`:

```go
package response

import (
	"net/http"

	apperrors "github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/errors"
	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, code int, message string) {
	c.JSON(statusCode, Envelope{
		Code:    code,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, apperrors.CodeBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, apperrors.CodeForbidden, message)
}

func NotImplemented(c *gin.Context) {
	Error(c, http.StatusNotImplemented, apperrors.CodeNotImplemented, "接口暂未实现")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/pkg/response -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pkg/errors/errors.go internal/pkg/response/response.go internal/pkg/response/response_test.go
git commit -m "feat: add response envelope helpers"
```

---

### Task 3: Add Config And Server Bootstrap

**Files:**
- Create: `backend/internal/config/config.go`
- Create: `backend/cmd/server/main.go`
- Test: `backend/internal/config/config_test.go`

- [ ] **Step 1: Write failing config test**

Create `backend/internal/config/config_test.go`:

```go
package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, 8080, cfg.Server.Port)
	require.Equal(t, "0.0.0.0", cfg.Server.Host)
	require.Equal(t, "ai-review", cfg.Auth.Issuer)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/config -run TestLoadDefaults -v
```

Expected: FAIL because config package does not exist.

- [ ] **Step 3: Implement config**

Create `backend/internal/config/config.go`:

```go
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Auth   AuthConfig   `mapstructure:"auth"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type AuthConfig struct {
	JWTSecret       string `mapstructure:"jwt_secret"`
	AccessTokenTTL  string `mapstructure:"access_token_ttl"`
	RefreshTokenTTL string `mapstructure:"refresh_token_ttl"`
	Issuer          string `mapstructure:"issuer"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("auth.jwt_secret", "dev-secret")
	v.SetDefault("auth.access_token_ttl", "30m")
	v.SetDefault("auth.refresh_token_ttl", "720h")
	v.SetDefault("auth.issuer", "ai-review")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
```

Create `backend/cmd/server/main.go` after router exists in Task 5. For now, create a placeholder that compiles after Task 5; if doing this task before router, skip `main.go` until Task 5.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/config -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: add config loading"
```

---

### Task 4: Add Middleware Skeleton

**Files:**
- Create: `backend/internal/middleware/auth.go`
- Create: `backend/internal/middleware/recovery.go`
- Create: `backend/internal/middleware/request_logger.go`
- Test: `backend/internal/middleware/auth_test.go`

- [ ] **Step 1: Write failing auth middleware tests**

Create `backend/internal/middleware/auth_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestJWTAuthRejectsMissingBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", JWTAuth(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthAllowsDevBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", JWTAuth(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer dev")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/middleware -run TestJWTAuth -v
```

Expected: FAIL because middleware does not exist.

- [ ] **Step 3: Implement middleware**

Create `backend/internal/middleware/auth.go`:

```go
package middleware

import (
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			response.Unauthorized(c, "未认证")
			c.Abort()
			return
		}
		if strings.TrimSpace(parts[1]) != "dev" {
			response.Unauthorized(c, "Token 无效")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("permission", permission)
		c.Next()
	}
}
```

Create `backend/internal/middleware/recovery.go`:

```go
package middleware

import "github.com/gin-gonic/gin"

func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}
```

Create `backend/internal/middleware/request_logger.go`:

```go
package middleware

import "github.com/gin-gonic/gin"

func RequestLogger() gin.HandlerFunc {
	return gin.Logger()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/middleware -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/middleware
git commit -m "feat: add middleware skeleton"
```

---

### Task 5: Register Public Routes And Server Bootstrap

**Files:**
- Create: `backend/internal/handler/stub.go`
- Create: `backend/internal/router/router.go`
- Create: `backend/cmd/server/main.go`
- Test: `backend/internal/router/router_contract_test.go`

- [ ] **Step 1: Write failing public route contract tests**

Create `backend/internal/router/router_contract_test.go`:

```go
package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPublicRoutesAreRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New()

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/v1/open/auth/login", `{}`},
		{http.MethodPost, "/api/v1/open/auth/refresh", `{}`},
		{http.MethodPost, "/api/v1/open/auth/logout", `{}`},
		{http.MethodPost, "/review/webhook", `{}`},
		{http.MethodGet, "/api/v1/open/system/info", ``},
		{http.MethodGet, "/api/v1/open/code-review-report", ``},
		{http.MethodGet, "/api/v1/open/analysis-report", ``},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		r.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "%s %s", tt.method, tt.path)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/router -run TestPublicRoutesAreRegistered -v
```

Expected: FAIL because router package does not exist.

- [ ] **Step 3: Implement public routes and server bootstrap**

Create `backend/internal/handler/stub.go`:

```go
package handler

import (
	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

func NotImplemented(c *gin.Context) {
	response.NotImplemented(c)
}

func SystemInfo(c *gin.Context) {
	response.Success(c, gin.H{
		"name":    "ai-review",
		"version": "0.0.0-dev",
	})
}
```

Create initial `backend/internal/router/router.go`:

```go
package router

import (
	"github.com/Lenoud/ai-review-gitlab/backend/internal/handler"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestLogger())

	open := r.Group("/api/v1/open")
	{
		open.POST("/auth/login", handler.NotImplemented)
		open.POST("/auth/refresh", handler.NotImplemented)
		open.POST("/auth/logout", handler.NotImplemented)
		open.GET("/system/info", handler.SystemInfo)
		open.GET("/code-review-report", handler.NotImplemented)
		open.GET("/analysis-report", handler.NotImplemented)
	}

	r.POST("/review/webhook", handler.NotImplemented)
	return r
}
```

Create `backend/cmd/server/main.go`:

```go
package main

import (
	"log"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/config"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	r := router.New()
	if err := r.Run(cfg.Server.Address()); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/router -run TestPublicRoutesAreRegistered -v
go test ./...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd internal/handler internal/router
git commit -m "feat: register public api contract routes"
```

---

### Task 6: Register Admin Contract Routes

**Files:**
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_contract_test.go`

- [ ] **Step 1: Add failing admin route contract test**

Append to `backend/internal/router/router_contract_test.go`:

```go
func TestAdminRoutesRequireAuthAndReturnNotImplementedWithDevToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New()

	routes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/v1/admin/project/create", `{}`},
		{http.MethodPost, "/api/v1/admin/project/batch-create", `{}`},
		{http.MethodPost, "/api/v1/admin/project/update", `{}`},
		{http.MethodGet, "/api/v1/admin/project/get", ``},
		{http.MethodPost, "/api/v1/admin/project/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/project/search", ``},
		{http.MethodPost, "/api/v1/admin/project/gitlab/remote-search", `{}`},
		{http.MethodPost, "/api/v1/admin/project/gitlab/group-search", `{}`},
		{http.MethodPost, "/api/v1/admin/project/web-urls/exists", `{}`},
		{http.MethodGet, "/api/v1/admin/project/review-prompt/get", ``},
		{http.MethodGet, "/api/v1/admin/project/review-prompt/default", ``},
		{http.MethodPost, "/api/v1/admin/project/review-prompt/update", `{}`},
		{http.MethodPost, "/api/v1/admin/project/review-prompt/delete", `{}`},
		{http.MethodPost, "/api/v1/admin/project/review-prompt/test", `{}`},
		{http.MethodGet, "/api/v1/admin/push-review-log/get", ``},
		{http.MethodGet, "/api/v1/admin/push-review-log/search", ``},
		{http.MethodPost, "/api/v1/admin/push-review-log/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/push-review-log/authors", ``},
		{http.MethodGet, "/api/v1/admin/push-review-log/project-names", ``},
		{http.MethodPost, "/api/v1/admin/push-review-log/generate-share-token/1", `{}`},
		{http.MethodGet, "/api/v1/admin/merge-request-review-log/get", ``},
		{http.MethodGet, "/api/v1/admin/merge-request-review-log/search", ``},
		{http.MethodPost, "/api/v1/admin/merge-request-review-log/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/merge-request-review-log/authors", ``},
		{http.MethodGet, "/api/v1/admin/merge-request-review-log/project-names", ``},
		{http.MethodPost, "/api/v1/admin/merge-request-review-log/generate-share-token/1", `{}`},
		{http.MethodPost, "/api/v1/admin/ai-review-trace/create", `{}`},
		{http.MethodGet, "/api/v1/admin/ai-review-trace/get", ``},
		{http.MethodPost, "/api/v1/admin/llm-model/create", `{}`},
		{http.MethodPost, "/api/v1/admin/llm-model/update", `{}`},
		{http.MethodGet, "/api/v1/admin/llm-model/get", ``},
		{http.MethodPost, "/api/v1/admin/llm-model/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/llm-model/search", ``},
		{http.MethodGet, "/api/v1/admin/llm-model/default", ``},
		{http.MethodPost, "/api/v1/admin/llm-model/set-default", `{}`},
		{http.MethodPost, "/api/v1/admin/llm-test/connection", `{}`},
		{http.MethodPost, "/api/v1/admin/im-robot/create", `{}`},
		{http.MethodPost, "/api/v1/admin/im-robot/update", `{}`},
		{http.MethodGet, "/api/v1/admin/im-robot/get", ``},
		{http.MethodPost, "/api/v1/admin/im-robot/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/im-robot/search", ``},
		{http.MethodGet, "/api/v1/admin/im-robot/list-enabled", ``},
		{http.MethodPost, "/api/v1/admin/im-robot/test-webhook", `{}`},
		{http.MethodPost, "/api/v1/admin/member-im-mapping/create", `{}`},
		{http.MethodPost, "/api/v1/admin/member-im-mapping/update", `{}`},
		{http.MethodGet, "/api/v1/admin/member-im-mapping/get", ``},
		{http.MethodPost, "/api/v1/admin/member-im-mapping/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/member-im-mapping/search", ``},
		{http.MethodPost, "/api/v1/admin/project-template/create", `{}`},
		{http.MethodPost, "/api/v1/admin/project-template/update", `{}`},
		{http.MethodGet, "/api/v1/admin/project-template/get", ``},
		{http.MethodGet, "/api/v1/admin/project-template/list", ``},
		{http.MethodPost, "/api/v1/admin/project-template/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/project-template/review-rule/list-by-template-id", ``},
		{http.MethodPost, "/api/v1/admin/project-template/review-rule/create", `{}`},
		{http.MethodPost, "/api/v1/admin/project-template/review-rule/update", `{}`},
		{http.MethodGet, "/api/v1/admin/project-template/review-rule/get", ``},
		{http.MethodPost, "/api/v1/admin/project-template/review-rule/delete", `{}`},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan/create", `{}`},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan/update", `{}`},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan/get", ``},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan/search", ``},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan-execution-log/test-run", `{}`},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan-execution-log/get", ``},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan-execution-log/search", ``},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan-execution-log/html-report/1", ``},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan-execution-log/generate-share-token/1", `{}`},
		{http.MethodPost, "/api/v1/admin/user/create", `{}`},
		{http.MethodPost, "/api/v1/admin/user/update", `{}`},
		{http.MethodGet, "/api/v1/admin/user/get", ``},
		{http.MethodGet, "/api/v1/admin/user/search", ``},
		{http.MethodGet, "/api/v1/admin/user/role-options", ``},
		{http.MethodGet, "/api/v1/admin/role/list", ``},
		{http.MethodPost, "/api/v1/admin/role/create", `{}`},
		{http.MethodPost, "/api/v1/admin/role/update", `{}`},
		{http.MethodGet, "/api/v1/admin/role/get", ``},
		{http.MethodPost, "/api/v1/admin/role/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/role/menu-permissions", ``},
		{http.MethodGet, "/api/v1/admin/auth/me", ``},
		{http.MethodGet, "/api/v1/admin/stats", ``},
		{http.MethodGet, "/api/v1/admin/member/commit-summary", ``},
		{http.MethodGet, "/api/v1/admin/sys-log/get", ``},
		{http.MethodGet, "/api/v1/admin/sys-log/search", ``},
		{http.MethodGet, "/api/v1/admin/system/info", ``},
		{http.MethodGet, "/api/v1/admin/system/config", ``},
		{http.MethodPost, "/api/v1/admin/system/config/base-url", `{}`},
		{http.MethodGet, "/api/v1/admin/review-log/get-share-token", ``},
	}

	for _, route := range routes {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code, "%s %s without token", route.method, route.path)

		w = httptest.NewRecorder()
		req = httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
		req.Header.Set("Authorization", "Bearer dev")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotImplemented, w.Code, "%s %s with token", route.method, route.path)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/router -run TestAdminRoutesRequireAuthAndReturnNotImplementedWithDevToken -v
```

Expected: FAIL because admin routes are not registered.

- [ ] **Step 3: Add route registration helper and admin routes**

Modify `backend/internal/router/router.go` to include:

```go
type routeDef struct {
	method string
	path   string
}

func registerRoutes(group *gin.RouterGroup, routes []routeDef) {
	for _, route := range routes {
		switch route.method {
		case http.MethodGet:
			group.GET(route.path, handler.NotImplemented)
		case http.MethodPost:
			group.POST(route.path, handler.NotImplemented)
		}
	}
}
```

Import `net/http`.

In `New()`, add:

```go
admin := r.Group("/api/v1/admin")
admin.Use(middleware.JWTAuth())
registerRoutes(admin, []routeDef{
	{http.MethodPost, "/project/create"},
	{http.MethodPost, "/project/batch-create"},
	// Add every route from the test list, without `/api/v1/admin` prefix.
})
```

Keep the route list in the same order as the spec/test to simplify review. For `/{logId}` style paths, use Gin `/:logId`.

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
go test ./internal/router -v
go test ./...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/router/router.go internal/router/router_contract_test.go
git commit -m "feat: register admin api contract routes"
```

---

### Task 7: Add DTO Placeholder Policy

**Files:**
- Create: `backend/internal/dto/README.md`
- Create: `backend/internal/dto/common.go`
- Test: `go test ./...`

- [ ] **Step 1: Add DTO policy document**

Create `backend/internal/dto/README.md`:

```markdown
# DTO Policy

DTOs define API request and response contracts. They are separate from GORM models.

In the API contract skeleton phase, handlers return `501 Not Implemented` and do not require full DTO coverage yet. When an endpoint is implemented, add its request/response DTOs in this package before writing service code.
```

- [ ] **Step 2: Add common DTO types**

Create `backend/internal/dto/common.go`:

```go
package dto

type PageRequest struct {
	Page int    `form:"page" json:"page"`
	Size int    `form:"size" json:"size"`
	Sort string `form:"sort" json:"sort"`
}

type PageResponse[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}
```

- [ ] **Step 3: Run tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/dto
git commit -m "chore: document dto policy"
```

---

### Task 8: Final Skeleton Verification

**Files:**
- Modify only if verification finds issues.

- [ ] **Step 1: Format**

Run:

```bash
make fmt
```

Expected: no output; files formatted.

- [ ] **Step 2: Test all packages**

Run:

```bash
make test
```

Expected: PASS.

- [ ] **Step 3: Start server**

Run:

```bash
go run ./cmd/server
```

Expected: server starts on `0.0.0.0:8080`.

- [ ] **Step 4: Smoke test in another terminal**

Run:

```bash
curl -i http://127.0.0.1:8080/api/v1/open/system/info
curl -i -H 'Authorization: Bearer dev' http://127.0.0.1:8080/api/v1/admin/project/search
```

Expected:
- public system info returns HTTP 200 and `code: 0`.
- admin project search returns HTTP 501 and `code: 50100`.

- [ ] **Step 5: Stop server and commit**

```bash
git status --short
git add .
git commit -m "test: verify api contract skeleton"
```

---

## Notes For Implementation

- Do not implement database logic in this plan.
- Do not return fake success for unimplemented endpoints.
- Keep all business endpoints wired to `handler.NotImplemented` until their real task is selected.
- The temporary `Bearer dev` auth is only for local contract testing. Replace it when implementing real auth.
- If route count changes in the spec, update the route contract test first, watch it fail, then update router registration.
