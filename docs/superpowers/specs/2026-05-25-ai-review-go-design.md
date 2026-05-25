# AI Code Review - Go 重写设计文档

## 背景

原项目是一个基于 Java Spring Boot 的 AI 代码审查平台（`com.mzfuture.aicr`），源码已永久丢失。现有 `app.jar`（106MB）已通过 jadx 反编译出 414 个 Java 文件。

本设计文档描述用 Go + Gin 重写的 V1 版本。

## V1 范围

### 做

- Webhook 接收 GitLab Push/Merge Request 事件，自动 AI 代码审查
- OpenAI 兼容协议通用 LLM 客户端
- 钉钉/飞书/企业微信 IM 通知
- HTML 审查报告（LLM 生成）+ 分享链接
- 定时分析计划（robfig/cron 替代 Quartz）
- 管理后台：项目、LLM 模型、审查日志、IM 机器人、模板、规则、分析计划
- JWT + RBAC 认证鉴权
- 用户/角色/权限管理
- 系统日志
- 成员提交统计

### 不做（后续迭代）

- Agent 对话（项目级 AI 问答）
- 深度审查（多轮探索模式）
- 多 Git 平台（GitHub/Gitee/Gitea）
- 商业授权（License 验证、项目上限、云端注册）
- 多 LLM Provider 独立客户端（10 种 → 1 种通用）

---

## 1. 目录结构

```
ai-review/
├── cmd/server/main.go           # 入口：加载配置、初始化 DB、注册路由、启动 HTTP
├── internal/
│   ├── config/                   # 配置结构体 + 加载逻辑
│   ├── middleware/               # JWT 认证、RBAC 权限、日志、CORS、Recovery
│   ├── router/                   # 路由注册，把 handler 绑到路由组
│   ├── handler/                  # HTTP handler，只做参数绑定和响应
│   ├── dto/                      # API 请求/响应结构体
│   ├── service/                  # 业务逻辑 + Repository 接口定义
│   ├── repository/               # 实现 service 定义的接口，直接操作 GORM
│   ├── model/                    # GORM 模型，纯数据库结构
│   ├── llm/                      # 外部适配器：OpenAI 兼容 LLM 客户端
│   ├── platform/                 # 外部适配器：GitLab API 交互
│   └── notify/                   # 外部适配器：钉钉/飞书/企业微信
├── migrations/                   # SQL 迁移脚本
├── config.yaml                   # 配置文件
├── go.mod
└── go.sum
```

### 依赖方向（单向）

```
handler → service → repository → model
               ↘ llm / platform / notify
```

- handler 不碰 repository
- service 不碰 GORM 细节（通过接口）
- model 和 dto 分离：model 是数据库结构，dto 是 API 契约

---

## 2. 数据库设计

### 15 张业务表

**核心审查**：

| 表名 | 关键字段 |
|------|------|
| `project` | id, name, description, webUrl, platform, accessToken, imEnabled, imRobotId, imAtMemberEnabled, imAtMemberScoreThreshold, aiReviewEnabled, templateId, extensions, reviewEventTypes, reviewPromptTemplate, htmlReportEnabled, deepReviewEnabled（字段保留但 V1 不使用） |
| `push_review_log` | id, projectId, projectName, author, authorIdentity, authorDisplayName, branch, commitMessages, commits(JSON), score, additions, deletions, lastCommitUrl, shareToken, shareTokenExpiresAt |
| `merge_request_review_log` | id, projectId, projectName, author, authorIdentity, authorDisplayName, sourceBranch, targetBranch, commitMessages, score, additions, deletions, lastCommitId, url, shareToken, shareTokenExpiresAt |
| `ai_review_trace` | id, reviewEventType, reviewEventId, prompt, response, provider, modelCode |
| `llm_model` | id, provider, modelCode, apiBaseUrl, apiKey, maxTokens, isDefault |

**通知**：

| 表名 | 关键字段 |
|------|------|
| `im_robot` | id, platform, name, webhookUrl, secret, enabled |
| `member_im_mapping` | id, gitUsername, platform, imUserId, displayName |

**模板**：

| 表名 | 关键字段 |
|------|------|
| `project_template` | id, name, description, extensions, reviewPromptTemplate |
| `project_template_review_rule` | id, templateId, name, description, globPatterns, content, priority, enabled |

**定时分析**：

| 表名 | 关键字段 |
|------|------|
| `project_analysis_plan` | id, projectId, name, prompt, cronExpression, enabled, imEnabled, imRobotId, htmlReportEnabled |
| `project_analysis_plan_execution_log` | id, planId(nullable), projectId(nullable), status, startedAt, completedAt, durationMs, resultContent, resultActions, shareToken, shareTokenExpiresAt, errorMessage, errorStack |

**RBAC**：

| 表名 | 关键字段 |
|------|------|
| `sys_user` | id, username, passwordHash, nickname, remark |
| `sys_role` | id, code, name, isSystem, remark |
| `sys_permission` | id, code, name, type, parentId, path, icon, sort, visible, remark |
| `sys_user_role` | id, userId, roleId |
| `sys_role_permission` | id, roleId, permissionId |

**系统**：

| 表名 | 关键字段 |
|------|------|
| `settings` | id, key, value(JSON) |
| `sys_log` | id, level, module, action, message, detail, errorStack |

### 不建的表

- `qrtz_*`（11 张）— Go 版用 robfig/cron 进程内调度
- `flyway_schema_history` — 用 golang-migrate 替代

### 表结构

直接复用 `init.sql` 中的字段定义，不做修改。初始数据也复用：

- `project_template`：4 条（Java/Vue/Go/全栈通用）
- `sys_role`：超级管理员
- `sys_permission`：15 条菜单权限
- `settings`：系统配置（去掉 License 和 Credential 相关）

---

## 3. 核心流程

### 3.1 Webhook 审查流程

```
GitLab Push/MR Event
       │
       ▼
  handler/webhook_handler.go
       │ 解析 payload，立即返回 200
       │ goroutine 异步处理（semaphore 控制并发）
       ▼
  service/review_service.go
       │
       ├── 1. 根据 webUrl 查 project 表获取配置
       ├── 2. platform/gitlab.go 调 GitLab API 获取 diff
       ├── 3. 按 extensions 过滤 diff 文件
       ├── 4. 组装 prompt（模板 + 规则 + diff + commit 信息）
       ├── 5. llm/client.go 调 OpenAI 兼容 API
       ├── 6. 解析评分，保存 review log
       ├── 7. platform/gitlab.go 回评到 GitLab（MR 评论/Commit 评论）
       ├── 8. 如果 htmlReportEnabled：llm 生成 HTML → 存文件系统
       └── 9. 如果 imEnabled：notify/sender.go 发 IM 通知
```

### 3.2 定时分析流程

```
robfig/cron 定时触发
       │
       ▼
  service/analysis_plan_service.go
       │
       ├── 1. 查 project_analysis_plan 获取计划配置
       ├── 2. 克隆/拉取项目代码
       ├── 3. 组装 prompt
       ├── 4. llm/client.go 调 LLM
       ├── 5. 保存 execution log
       ├── 6. 如果 htmlReportEnabled：生成 HTML 报告
       └── 7. 如果 imEnabled：发 IM 通知
```

### 3.3 LLM 调用

```
llm/client.go
  ├── 统一 OpenAI 兼容客户端（POST /chat/completions）
  ├── 配置来自 llm_model 表（apiBaseUrl, apiKey, modelCode, maxTokens）
  ├── 每次调用记录到 ai_review_trace 表
  └── 错误分类（网络错误、限流、Token 超限等）
```

### 3.4 HTML 报告

```
html_report_service.go
  ├── 把审查结果 + 设计要求拼成 prompt 发给 LLM
  ├── LLM 输出完整 HTML（内联 CSS + SVG，自包含页面）
  ├── 存到文件系统：data/code-review-html/{projectId}/{logType}/{logId}.html
  └── 分享 Token：UUID 生成，7 天过期，存数据库 share_token 字段
```

### 3.5 依赖注入

Go 用构造函数注入，在 main.go 中手动组装：

```go
db := initDB()
projectRepo := repository.NewProjectRepo(db)
llmClient := llm.NewClient(...)
gitlabAdapter := platform.NewGitlabAdapter(...)
notifySender := notify.NewSender(...)
reviewService := service.NewReviewService(projectRepo, llmClient, gitlabAdapter, notifySender)
reviewHandler := handler.NewReviewHandler(reviewService)
```

---

## 4. API 接口清单（约 80 个）

分页请求参数统一约定：`page`（从 0 开始）、`size`（每页数量，默认 20）、`sort`（格式 `field,direction`，如 `createdAt,desc`）。

### 4.1 开放接口（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/open/auth/login` | 登录 |
| POST | `/api/v1/open/auth/refresh` | 刷新 Token |
| POST | `/api/v1/open/auth/logout` | 登出 |
| POST | `/review/webhook` | 接收 GitLab Webhook（兼容原版路径） |
| GET | `/api/v1/open/system/info` | 公开系统信息 |
| GET | `/api/v1/open/code-review-report` | 分享链接查看代码审查报告 |
| GET | `/api/v1/open/analysis-report` | 分享链接查看分析报告 |

### 4.2 项目管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/project/create` | 创建项目 |
| POST | `/api/v1/admin/project/batch-create` | 批量创建项目 |
| POST | `/api/v1/admin/project/update` | 更新项目 |
| GET | `/api/v1/admin/project/get` | 查询项目 |
| POST | `/api/v1/admin/project/delete` | 删除项目 |
| GET | `/api/v1/admin/project/search` | 搜索项目 |
| POST | `/api/v1/admin/project/gitlab/remote-search` | GitLab 远程项目列表 |
| POST | `/api/v1/admin/project/gitlab/group-search` | GitLab 分组搜索 |
| POST | `/api/v1/admin/project/web-urls/exists` | 检测 URL 是否已存在 |
| GET | `/api/v1/admin/project/review-prompt/get` | 获取项目 Review 提示词 |
| GET | `/api/v1/admin/project/review-prompt/default` | 获取默认 Review 提示词 |
| POST | `/api/v1/admin/project/review-prompt/update` | 更新 Review 提示词 |
| POST | `/api/v1/admin/project/review-prompt/delete` | 删除项目自定义提示词 |
| POST | `/api/v1/admin/project/review-prompt/test` | 测试提示词渲染 |

### 4.3 审查日志

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/push-review-log/get` | 查询 Push 审查日志 |
| GET | `/api/v1/admin/push-review-log/search` | 搜索 Push 审查日志 |
| POST | `/api/v1/admin/push-review-log/delete` | 删除 |
| GET | `/api/v1/admin/push-review-log/authors` | 提交者列表 |
| GET | `/api/v1/admin/push-review-log/project-names` | 项目名称列表 |
| POST | `/api/v1/admin/push-review-log/generate-share-token/{logId}` | 生成分享 Token |
| GET | `/api/v1/admin/merge-request-review-log/get` | 查询 MR 审查日志 |
| GET | `/api/v1/admin/merge-request-review-log/search` | 搜索 MR 审查日志 |
| POST | `/api/v1/admin/merge-request-review-log/delete` | 删除 |
| GET | `/api/v1/admin/merge-request-review-log/authors` | 提交者列表 |
| GET | `/api/v1/admin/merge-request-review-log/project-names` | 项目名称列表 |
| POST | `/api/v1/admin/merge-request-review-log/generate-share-token/{logId}` | 生成分享 Token |

### 4.4 AI 审查追踪

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/ai-review-trace/create` | 创建追踪记录 |
| GET | `/api/v1/admin/ai-review-trace/get` | 查询追踪记录 |

### 4.4 LLM 模型管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/llm-model/create` | 创建 |
| POST | `/api/v1/admin/llm-model/update` | 更新 |
| GET | `/api/v1/admin/llm-model/get` | 查询 |
| POST | `/api/v1/admin/llm-model/delete` | 删除 |
| GET | `/api/v1/admin/llm-model/search` | 搜索 |
| GET | `/api/v1/admin/llm-model/default` | 获取默认模型 |
| POST | `/api/v1/admin/llm-model/set-default` | 设置默认模型 |
| POST | `/api/v1/admin/llm-test/connection` | 测试连接 |

### 4.5 IM 机器人管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/im-robot/create` | 创建 |
| POST | `/api/v1/admin/im-robot/update` | 更新 |
| GET | `/api/v1/admin/im-robot/get` | 查询 |
| POST | `/api/v1/admin/im-robot/delete` | 删除 |
| GET | `/api/v1/admin/im-robot/search` | 搜索 |
| GET | `/api/v1/admin/im-robot/list-enabled` | 启用列表 |
| POST | `/api/v1/admin/im-robot/test-webhook` | 测试连接 |

### 4.6 成员 IM 映射

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/member-im-mapping/create` | 创建 |
| POST | `/api/v1/admin/member-im-mapping/update` | 更新 |
| GET | `/api/v1/admin/member-im-mapping/get` | 查询 |
| POST | `/api/v1/admin/member-im-mapping/delete` | 删除 |
| GET | `/api/v1/admin/member-im-mapping/search` | 搜索 |

### 4.7 项目模板

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/project-template/create` | 创建 |
| POST | `/api/v1/admin/project-template/update` | 更新 |
| GET | `/api/v1/admin/project-template/get` | 查询 |
| GET | `/api/v1/admin/project-template/list` | 全部列表 |
| POST | `/api/v1/admin/project-template/delete` | 删除 |

### 4.8 审查规则

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/project-template/review-rule/list-by-template-id` | 按模板列出 |
| POST | `/api/v1/admin/project-template/review-rule/create` | 创建 |
| POST | `/api/v1/admin/project-template/review-rule/update` | 更新 |
| GET | `/api/v1/admin/project-template/review-rule/get` | 查询 |
| POST | `/api/v1/admin/project-template/review-rule/delete` | 删除 |

### 4.9 定时分析计划

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/project-analysis-plan/create` | 创建 |
| POST | `/api/v1/admin/project-analysis-plan/update` | 更新 |
| GET | `/api/v1/admin/project-analysis-plan/get` | 查询 |
| POST | `/api/v1/admin/project-analysis-plan/delete` | 删除 |
| GET | `/api/v1/admin/project-analysis-plan/search` | 搜索 |
| POST | `/api/v1/admin/project-analysis-plan-execution-log/test-run` | 手动执行 |
| GET | `/api/v1/admin/project-analysis-plan-execution-log/get` | 查询执行日志 |
| GET | `/api/v1/admin/project-analysis-plan-execution-log/search` | 搜索执行日志 |
| GET | `/api/v1/admin/project-analysis-plan-execution-log/html-report/{logId}` | 查看 HTML 报告 |
| POST | `/api/v1/admin/project-analysis-plan-execution-log/generate-share-token/{logId}` | 生成分享 Token |

### 4.10 用户 / 角色 / 权限

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/user/create` | 创建用户 |
| POST | `/api/v1/admin/user/update` | 更新用户 |
| GET | `/api/v1/admin/user/get` | 查询 |
| GET | `/api/v1/admin/user/search` | 搜索 |
| GET | `/api/v1/admin/user/role-options` | 角色选项 |
| GET | `/api/v1/admin/role/list` | 角色列表 |
| POST | `/api/v1/admin/role/create` | 创建角色 |
| POST | `/api/v1/admin/role/update` | 更新角色 |
| GET | `/api/v1/admin/role/get` | 查询角色 |
| POST | `/api/v1/admin/role/delete` | 删除角色 |
| GET | `/api/v1/admin/role/menu-permissions` | 菜单权限列表 |
| GET | `/api/v1/admin/auth/me` | 当前用户 |

### 4.12 统计 & 系统

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/stats` | 统计数据 |
| GET | `/api/v1/admin/member/commit-summary` | 成员提交统计 |
| GET | `/api/v1/admin/sys-log/get` | 查询单条系统日志 |
| GET | `/api/v1/admin/sys-log/search` | 系统日志搜索 |
| GET | `/api/v1/admin/system/info` | 系统信息 |
| GET | `/api/v1/admin/system/config` | 系统配置 |
| POST | `/api/v1/admin/system/config/base-url` | 更新访问地址 |
| GET | `/api/v1/admin/review-log/get-share-token` | 查询审查日志分享 Token |

---

## 5. 配置文件

```yaml
server:
  port: 8080
  base_url: "http://localhost:8080"

database:
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: ""
  dbname: "codereview"
  max_idle_conns: 10
  max_open_conns: 100

auth:
  jwt_secret: "your-secret-key"
  access_token_ttl: "30m"
  refresh_token_ttl: "720h"
  issuer: "ai-review"
  admin_username: "admin"
  admin_password: "admin123"

data_dir: "data"

log:
  level: "info"
  format: "json"
```

应急管理员只在首次启动时（数据库无用户时）自动创建。`data_dir` 相对路径基于二进制所在目录。

---

## 6. 中间件

### 请求链

```
请求 → Logger → Recovery → CORS → JWT认证 → RBAC权限 → Handler
```

### 路由分组

```
公开路由组（无 JWT）
  ├── /api/v1/open/*
  └── /review/webhook

认证路由组（需要 JWT）
  ├── /api/v1/admin/auth/me
  └── 大部分查询接口

RBAC 路由组（需要特定权限）
  ├── /api/v1/admin/user/*   → user_management
  └── /api/v1/admin/role/*   → role_management
```

### 统一响应格式

成功：

```json
{ "code": 0, "data": { ... } }
```

分页：

```json
{ "code": 0, "data": { "items": [...], "total": 100, "page": 0, "size": 20 } }
```

失败：

```json
{ "code": 40101, "message": "Token 已过期" }
```

错误码分段：

| 范围 | 含义 |
|------|------|
| 40000-40099 | 参数校验错误 |
| 40100-40199 | 认证错误 |
| 40300-40399 | 权限不足 |
| 40400-40499 | 资源不存在 |
| 50000-50099 | 服务端内部错误 |

---

## 7. 技术决策

| 项 | 决策 | 理由 |
|------|------|------|
| Webhook 异步 | goroutine + semaphore.Weighted | 防止大量 Webhook 同时打满 LLM API |
| 定时任务 | robfig/cron/v3 | 进程内调度，从数据库动态加载计划 |
| Cron 兼容性 | 代码中做 Quartz → 标准 cron 转换 | init.sql 中 cronExpression 注释为 Quartz 格式。robfig/cron 不支持 Quartz 的 `?` 和 `周 1-7`。在加载时自动转换：`?` → `*`，周字段 `1-7` → `0-6` |
| LLM 调用 | 非流式，单次 /chat/completions | V1 只做审查和分析，不需要流式 |
| 密码存储 | golang.org/x/crypto/bcrypt | 与原版一致 |
| 数据库迁移 | golang-migrate/migrate | 应用启动时自动执行 |
| HTML 报告 | LLM 生成完整 HTML，存文件系统 | 与原版一致，无模板文件 |
| HTML 报告存储 | 定义 HtmlReportStorage 接口 | V1 实现为本地文件系统，后续可切换到 S3/MinIO。接口隔离实现 |
| 分享 Token | UUID，7 天过期，存数据库 | 与原版一致 |
| 项目代码拉取 | 调用系统 git 命令 | 比 go-git 更稳定，支持所有 Git 特性。代码存储在 `{data_dir}/projects/{projectId}/`。大型仓库不特殊处理，自然克隆 |
| Graceful shutdown | 监听 SIGTERM，等待进行中的 Webhook 和 cron 任务完成 | 使用 `sync.WaitGroup` 跟踪进行中的 goroutine，shutdown 时 Wait 或超时强制退出 |
| 应急管理员密码 | config.yaml 中明文 + 支持环境变量覆盖 | 环境变量 `ADMIN_PASSWORD` 优先于配置文件，避免密码意外提交到版本库 |

### Go 依赖（11 个）

```
github.com/gin-gonic/gin            # HTTP 框架
gorm.io/gorm                         # ORM
gorm.io/driver/mysql                 # MySQL 驱动
github.com/golang-jwt/jwt/v5         # JWT
golang.org/x/crypto                  # bcrypt
github.com/robfig/cron/v3            # 定时任务
github.com/golang-migrate/migrate/v4 # 数据库迁移
github.com/go-resty/resty/v2         # HTTP 客户端
github.com/spf13/viper               # 配置文件读取
go.uber.org/zap                      # 日志
github.com/google/uuid               # UUID 生成
```
