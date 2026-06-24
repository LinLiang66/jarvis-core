# jarvis

前后端分离的管理后台基础框架：**系统管理** + **开放平台**，适合作为业务系统的起点快速扩展。

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
