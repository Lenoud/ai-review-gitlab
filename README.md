# ai-review-gitlab

基于 AI 的 GitLab 代码审阅工具。

## 功能

- 对接 GitLab MR Webhook，自动触发代码审阅
- 基于 AI 模型对代码变更进行智能分析，给出审阅意见
- 支持审阅结果评论回写到 GitLab MR

## 技术栈

- Go
- GitLab API
- AI 模型集成

## 项目结构

```
backend/      # 后端服务
  cmd/        # 入口
  internal/   # 业务逻辑
  migrations/ # 数据库迁移
docs/         # 文档
```

## 快速开始

```bash
cp backend/config.example.yaml backend/config.yaml
# 编辑 config.yaml 填入 GitLab 和 AI 模型配置
cd backend && go run cmd/main.go
```
