# jarvis

jarvis 是一套前后端分离的管理后台基础框架，面向需要快速搭建企业级后台、又希望保留统一对外 API 能力的团队。项目从完整 Admin 体系中精简而来，聚焦两大核心：系统管理与开放平台，去掉 AI、语音、外呼等业务模块，代码结构清晰，适合作为二次开发的起点或直接开源发布。

后端采用 Go + Gin + GORM，提供 REST API；支持 MySQL（生产推荐）与 SQLite（本地开发）；可选 Redis 用于登录态与开放平台会话共享。认证基于 JWT，内置 RBAC：用户、角色、菜单、字典等管理能力开箱即用，菜单驱动前端动态路由，权限可精细到按钮级。

前端基于 Vue 3、TypeScript、Vite、Element Plus 与 gi-component，提供工作台、系统管理、开放平台等页面，支持主题切换、标签页、响应式布局，开发体验与上游 gi-element-plus-admin 一脉相承。

开放平台是 jarvis 的差异化能力：对外提供统一网关入口，第三方应用通过 AppID、签名与 3DES 加密调用注册的 Action；管理端支持应用管理、接口同步、在线文档、调用统计与配额控制。仓库附带 Go / Python / Java SDK 示例（Echo 演示），文档中说明了网关集群与调用方集群的 Redis 会话共享、握手加锁等生产实践。

项目包含完整 docs/ 文档（快速开始、架构、API、部署、开发指南）、Docker 后端部署模板，以及 Cursor Skills 等工程化辅助。无论是搭建内部管理系统、Saas 运营后台，还是在统一网关之上扩展行业 Action，jarvis 都可以作为稳定、可扩展的基础底座。

## 特性

- RBAC 权限、动态菜单路由、JWT 登录
- 用户 / 角色 / 菜单 / 字典管理
- 开放平台：统一网关、应用管理、接口文档、调用统计
- Vue 3 + Element Plus + gi-component 管理界面
- Go + Gin + GORM，MySQL / SQLite 双模式

## 快速开始

```powershell
# 后端
cd backend && copy .env.example .env && go run ./cmd/server

# 前端（新终端）
cd frontend/web && pnpm install && pnpm dev
```

访问 http://localhost:5050 ，默认账号 `admin` / `123456`。

## 文档

完整文档见 **[docs/](docs/README.md)**：

| 文档 | 说明 |
|------|------|
| [快速开始](docs/getting-started.md) | 环境、启动、常见问题 |
| [架构说明](docs/architecture.md) | 模块与认证流程 |
| [API 参考](docs/api-reference.md) | REST 接口一览 |
| [开放平台](docs/openplatform.md) | 网关协议与接入 |
| [部署指南](docs/deployment.md) | Docker 与生产部署 |
| [开发指南](docs/development.md) | 二次开发说明 |

## 项目结构

```text
.
├── backend/           # Go API
├── frontend/web/      # Vue 3 管理后台
├── docker/            # 后端 Docker
├── examples/          # 开放平台 SDK 示例
└── docs/              # 项目文档
```

## 环境要求

Go 1.21+ · Node.js 18+ · pnpm · MySQL（推荐）· Redis（可选）

## Git 远程

| 远程 | 地址 |
|------|------|
| origin | https://gitcode.com/LinLiang/jarvis-core.git |
| upstream | https://github.com/lin-97/gi-element-plus-admin.git |
