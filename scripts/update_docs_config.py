# -*- coding: utf-8 -*-
"""Update jarvis-core deployment configs (.env, docker-compose).

Project documentation: scripts/write_project_docs.py
Agent skill: .cursor/skills/docs-config/SKILL.md
"""
from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

BACKEND_ENV = """SERVER_ADDR=:8000
LOG_LEVEL=info

# SQLite (default when MySQL is not configured)
DB_PATH=./data/app.db

# MySQL (used when MYSQL_HOST and MYSQL_DATABASE are set)
MYSQL_HOST=
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=
MYSQL_DATABASE=jarvis_core
MYSQL_CHARSET=utf8mb4
MYSQL_SHOW_SQL=false
MYSQL_LOG_LEVEL=2
MYSQL_MAX_OPEN_CONNS=20
MYSQL_MAX_IDLE_CONNS=10

# Redis (optional; login token cache; falls back to JWT-only when unavailable)
REDIS_ENABLE=true
REDIS_REQUIRED=false
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_READ_TIMEOUT=3s

JWT_SECRET=jarvis-core-dev-secret
JWT_EXPIRE_HOURS=24
JWT_REFRESH_DAYS=7

UPLOAD_DIR=./data/uploads
STATIC_URL_PREFIX=/static
PUBLIC_BASE_URL=http://127.0.0.1:8000

# Image upload smart compression (local + OSS)
IMAGE_COMPRESS_ENABLE=true
IMAGE_COMPRESS_MAX_DIM=1920
IMAGE_COMPRESS_QUALITY=85
IMAGE_COMPRESS_MIN_BYTES=102400
IMAGE_COMPRESS_MAX_INPUT=20971520

# Set false in production when schema is stable to skip AutoMigrate on boot
DB_AUTO_MIGRATE=true

# Scheduler (independent scheduler-server; jarvis acts as Worker + BFF proxy)
SCHEDULER_ENABLE=false
SCHEDULER_SERVER_URL=http://127.0.0.1:9000
SCHEDULER_ADMIN_TOKEN=sched-admin-dev
SCHEDULER_WORKER_TOKEN=sched-worker-dev
SCHEDULER_INSTANCE_ID=
# Match scheduler POLL_TIMEOUT_SEC; empty-poll backoff (ms)
SCHEDULER_POLL_TIMEOUT_SEC=30
SCHEDULER_POLL_EMPTY_BACKOFF_MS=200
# When dispatch gate is closed, worker idle recheck interval (sec)
SCHEDULER_POLL_IDLE_SEC=30
"""

SCHEDULER_ENV = """SERVER_ADDR=:9000

# MySQL (scheduler uses its own database jarvis_scheduler)
# Legacy override: set MYSQL_DSN to skip component vars below
# MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/jarvis_scheduler?charset=utf8mb4&parseTime=True&loc=Local
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=root
MYSQL_DATABASE=jarvis_scheduler
MYSQL_CHARSET=utf8mb4
MYSQL_SHOW_SQL=false
MYSQL_LOG_LEVEL=2
MYSQL_MAX_OPEN_CONNS=20
MYSQL_MAX_IDLE_CONNS=10

# Redis (required; task queue, worker heartbeat, distributed locks)
REDIS_ENABLE=true
REDIS_REQUIRED=true
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_READ_TIMEOUT=3s

# API tokens (must match backend SCHEDULER_ADMIN_TOKEN / SCHEDULER_WORKER_TOKEN when integrated)
ADMIN_TOKEN=sched-admin-dev
WORKER_TOKEN=sched-worker-dev

# Engine tuning (seconds / milliseconds)
POLL_TIMEOUT_SEC=30
POLL_INTERVAL_MS=1000
WORKER_TTL_SEC=90
RUNNING_LOCK_TTL_SEC=3600
# Instance scanner: reclaim stale claims / fail execution timeouts / discard when no worker
INSTANCE_CLAIM_TIMEOUT_SEC=120
INSTANCE_SCAN_INTERVAL_SEC=30
"""

DOCKER_ENV = """# Host port mapping: host:666 -> container:8000
BACKEND_PORT=666
PUBLIC_BASE_URL=http://localhost:666

SERVER_ADDR=:8000
LOG_LEVEL=info

DB_PATH=/app/data/app.db
UPLOAD_DIR=/app/data/uploads
STATIC_URL_PREFIX=/static

IMAGE_COMPRESS_ENABLE=true
IMAGE_COMPRESS_MAX_DIM=1920
IMAGE_COMPRESS_QUALITY=85
IMAGE_COMPRESS_MIN_BYTES=102400
IMAGE_COMPRESS_MAX_INPUT=20971520

DB_AUTO_MIGRATE=true

# MySQL (recommended for production)
MYSQL_HOST=
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=
MYSQL_DATABASE=jarvis_core
MYSQL_CHARSET=utf8mb4
MYSQL_SHOW_SQL=false
MYSQL_LOG_LEVEL=2
MYSQL_MAX_OPEN_CONNS=20
MYSQL_MAX_IDLE_CONNS=10

# Redis (optional)
REDIS_ENABLE=true
REDIS_REQUIRED=false
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_READ_TIMEOUT=3s

JWT_SECRET=jarvis-core-change-me-in-production
JWT_EXPIRE_HOURS=24

# Scheduler (independent service; see scheduler/.env.example for local run)
SCHEDULER_PORT=9000
SCHEDULER_MYSQL_HOST=host.docker.internal
SCHEDULER_MYSQL_PORT=3306
SCHEDULER_MYSQL_USER=root
SCHEDULER_MYSQL_PASSWORD=root
SCHEDULER_MYSQL_DATABASE=jarvis_scheduler
# Legacy DSN override for scheduler container (optional; maps to MYSQL_DSN inside container)
# SCHEDULER_MYSQL_DSN=root:root@tcp(host.docker.internal:3306)/jarvis_scheduler?charset=utf8mb4&parseTime=True&loc=Local
SCHEDULER_ADMIN_TOKEN=sched-admin-dev
SCHEDULER_WORKER_TOKEN=sched-worker-dev
SCHEDULER_SERVER_URL=http://scheduler:9000
SCHEDULER_ENABLE=true
SCHEDULER_POLL_TIMEOUT_SEC=30
SCHEDULER_POLL_EMPTY_BACKOFF_MS=200
SCHEDULER_POLL_IDLE_SEC=30

# Scheduler engine tuning (passed to scheduler container)
POLL_TIMEOUT_SEC=30
POLL_INTERVAL_MS=1000
WORKER_TTL_SEC=90
RUNNING_LOCK_TTL_SEC=3600
INSTANCE_CLAIM_TIMEOUT_SEC=120
INSTANCE_SCAN_INTERVAL_SEC=30
"""

BACKEND_README = """# jarvis-core 后端

Go 语言 REST API 服务，基于 **Gin + GORM + JWT**，提供系统管理与开放平台基础能力。

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 / 框架 | Go 1.21+、Gin |
| ORM | GORM |
| 数据库 | MySQL（推荐）、SQLite（开发降级） |
| 缓存 | Redis（可选，登录 token 缓存） |
| 认证 | JWT |

## 目录结构

```text
backend/
├── cmd/server/main.go       # 服务入口
├── internal/
│   ├── config/              # 环境变量配置
│   ├── router/              # 路由注册
│   ├── handler/             # HTTP 处理器（auth、system、openplatform）
│   ├── middleware/          # JWT、超管等中间件
│   ├── service/openplatform/# 开放平台网关与业务
│   ├── store/               # 数据访问层
│   ├── model/               # 数据模型
│   └── database/            # 数据库初始化与种子数据
├── sql/                     # 可选 SQL 补丁
├── .env.example             # 环境变量示例
└── go.mod
```

## 快速开始

### 1. 配置

```powershell
cd backend
copy .env.example .env
# 编辑 .env；生产环境请修改 JWT_SECRET
# 配置 MYSQL_HOST + MYSQL_DATABASE 后自动切换 MySQL
```

| 变量 | 说明 | 默认 |
|------|------|------|
| `SERVER_ADDR` | 监听地址 | `:8000` |
| `DB_PATH` | SQLite 路径 | `./data/app.db` |
| `MYSQL_HOST` / `MYSQL_DATABASE` | MySQL 连接 | — |
| `REDIS_ENABLE` | 是否启用 Redis | `true` |
| `REDIS_REQUIRED` | Redis 不可用时是否拒绝启动 | `false` |
| `JWT_SECRET` | JWT 签名密钥 | `jarvis-core-dev-secret` |
| `JWT_EXPIRE_HOURS` | Token 有效期（小时） | `24` |

### 2. 启动

```powershell
go run ./cmd/server
# 或 Windows: run_win.bat
# 或 Linux:   ./run_linux.sh
```

健康检查：`GET http://localhost:8000/health`

### 3. 测试

```powershell
go test ./...
```

## API 概览

业务接口前缀 `/api/v1`，除登录外需 `Authorization: Bearer <token>`。

| 模块 | 路径前缀 | 说明 |
|------|----------|------|
| 认证 | `/api/v1/auth` | 登录、刷新 token |
| 系统 | `/api/v1/system/*` | 用户、角色、菜单、字典 |
| 开放平台网关 | `/api/v1/open/gateway` | 对外网关（form-urlencoded） |
| 开放平台管理 | `/api/v1/open-app/*` | 应用、接口、统计、文档 |

## 数据库

- 配置 `MYSQL_HOST` + `MYSQL_DATABASE` 后优先 MySQL；否则使用 SQLite。
- 首次启动自动迁移并写入种子数据（默认管理员 `admin` / `123456`）。

## Docker 部署

见项目根目录 [docker/README.md](../docker/README.md)。

- 模块名：`jarvis-core/backend`
- Redis 连接失败时默认降级为纯 JWT（`REDIS_REQUIRED=false`）
"""

FRONTEND_README = """# jarvis-core 前端

基于 Vue 3 + TypeScript + Element Plus 的管理后台，配合 jarvis-core 后端使用。

## 技术栈

- Vue 3、TypeScript、Vite 7
- Element Plus、gi-component、Pinia、Vue Router 4

## 功能

- RBAC 权限、动态路由、登录态管理
- 系统管理：用户、角色、菜单、字典
- 开放平台：应用、接口、文档、调用统计
- 工作台、主题切换、标签页

## 快速开始

```bash
pnpm install
pnpm dev
```

默认访问 http://localhost:5050 ，API 代理至 `http://localhost:8000`（需先启动后端）。

## 环境变量

| 变量 | 说明 | 默认 |
|------|------|------|
| `VITE_APP_TITLE` | 浏览器标题 | Jarvis |
| `VITE_BASE` | 部署基础路径 | `/` |
| `VITE_API_BASE_URL` | API 前缀 | `/api/v1` |

## 构建

```bash
pnpm build
pnpm preview
```
"""

DOCKER_README = """# Docker 部署

部署 **scheduler-server**（`:9000`）与 **jarvis 后端 API**（容器 `:8000`，宿主机默认 `:666`）；MySQL、Redis 使用外部服务。

## 使用步骤

```powershell
cd docker
copy .env.example .env
# 编辑 .env：MYSQL_*、SCHEDULER_MYSQL_*、REDIS_*、JWT_SECRET、SCHEDULER_* tokens
docker compose up -d --build
```

| 服务 | 健康检查 | 默认宿主机端口 |
|------|----------|----------------|
| jarvis 后端 | `GET http://localhost:666/health` | `666`（`BACKEND_PORT`） |
| scheduler-server | `GET http://localhost:9000/health` | `9000`（`SCHEDULER_PORT`） |

## 说明

- **部署拓扑**：1 个 scheduler-server + N 个 jarvis 后端 Worker；compose 默认各 1 副本
- scheduler 使用独立库 **`jarvis_scheduler`**（与后端 `jarvis_core` 分离）
- backend 容器通过 `SCHEDULER_SERVER_URL=http://scheduler:9000` 连接调度服务，并内嵌 Worker 客户端
- 镜像内默认 SQLite 数据目录：`/app/data`（Docker volume `backend-data`）
- 生产环境建议配置远程 MySQL，并修改 `JWT_SECRET` 与 `SCHEDULER_*_TOKEN`
- 前端需单独构建部署，或将 `frontend/web/dist` 交由 Nginx 托管

详见 [docs/deployment.md](../docs/deployment.md) 与 [docs/scheduler.md](../docs/scheduler.md)。
"""

ROOT_README = """# jarvis-core

前后端分离的管理后台基础框架：保留系统管理能力（用户、角色、菜单、字典）与开放平台基础能力（应用、接口、文档、统计），便于在此基础上快速搭建业务系统。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go、Gin、GORM、JWT、MySQL/SQLite、Redis（可选） |
| 前端 | Vue 3、Vite、TypeScript、Element Plus、gi-component |

## 项目结构

```text
.
├── backend/           # Go API
├── frontend/web/      # Vue 3 管理后台
├── docker/            # 后端 Docker 部署
├── examples/          # 开放平台 SDK 示例
└── README.md
```

## Git 远程

| 远程 | 地址 | 用途 |
|------|------|------|
| origin | https://gitcode.com/LinLiang/jarvis-core.git | 提交代码（master） |
| upstream | https://github.com/lin-97/gi-element-plus-admin.git | 跟踪上游框架 |

```bash
git fetch upstream
git merge upstream/main
```

## 本地开发

### 后端

```powershell
cd backend
copy .env.example .env
go run ./cmd/server
```

默认 `:8000`，健康检查 `GET /health`。

### 前端

```powershell
cd frontend/web
pnpm install
pnpm dev
```

默认 `:5050`，API 代理至后端。

### 验证

```powershell
cd backend && go test ./...
cd frontend/web && pnpm build
```

## Docker 部署（仅后端）

```powershell
cd docker
copy .env.example .env
docker compose up -d --build
```

详见 [docker/README.md](docker/README.md)。

## 核心功能

| 功能 | 说明 |
|------|------|
| 系统管理 | 用户、角色、菜单、字典 |
| 开放平台 | 网关、应用、接口文档、调用统计 |
| 工作台 | 基础仪表盘 |

默认管理员：`admin` / `123456`

## 文档

- [backend/README.md](backend/README.md) — 后端 API 与配置
- [frontend/web/README.md](frontend/web/README.md) — 前端开发
- [docker/README.md](docker/README.md) — 容器部署

## 环境要求

- Go 1.21+、Node.js 18+、pnpm
- MySQL（生产推荐）、Redis（可选）
"""

DOCKER_COMPOSE = """# jarvis-core 部署：scheduler-server (:9000) + jarvis 后端 API (:8000；宿主机默认 :666)
#
# 使用：
#   cd docker
#   copy .env.example .env
#   docker compose up -d --build
#
# 访问：
#   http://localhost:666/health        (backend)
#   http://localhost:9000/health       (scheduler)

services:
  # Standalone local dev: copy scheduler/.env.example to scheduler/.env, then `cd scheduler && go run ./cmd/server`
  # Docker env vars below (or override via docker/.env: SCHEDULER_*, REDIS_*, POLL_*, INSTANCE_*)
  scheduler:
    build:
      context: ../scheduler
      dockerfile: Dockerfile
      args:
        DEBIAN_MIRROR: ${DEBIAN_MIRROR:-mirrors.aliyun.com}
    image: jarvis-core-scheduler:latest
    container_name: jarvis-core-scheduler
    restart: unless-stopped
    ports:
      - "${SCHEDULER_PORT:-9000}:9000"
    environment:
      SERVER_ADDR: ":9000"
      MYSQL_DSN: ${SCHEDULER_MYSQL_DSN:-}
      MYSQL_HOST: ${SCHEDULER_MYSQL_HOST:-host.docker.internal}
      MYSQL_PORT: ${SCHEDULER_MYSQL_PORT:-3306}
      MYSQL_USER: ${SCHEDULER_MYSQL_USER:-root}
      MYSQL_PASSWORD: ${SCHEDULER_MYSQL_PASSWORD:-root}
      MYSQL_DATABASE: ${SCHEDULER_MYSQL_DATABASE:-jarvis_scheduler}
      MYSQL_CHARSET: ${SCHEDULER_MYSQL_CHARSET:-utf8mb4}
      REDIS_ENABLE: ${REDIS_ENABLE:-true}
      REDIS_REQUIRED: ${REDIS_REQUIRED:-true}
      REDIS_ADDR: ${REDIS_ADDR:-host.docker.internal:6379}
      REDIS_PASSWORD: ${REDIS_PASSWORD:-}
      REDIS_DB: ${REDIS_DB:-0}
      REDIS_POOL_SIZE: ${REDIS_POOL_SIZE:-10}
      ADMIN_TOKEN: ${SCHEDULER_ADMIN_TOKEN:-sched-admin-dev}
      WORKER_TOKEN: ${SCHEDULER_WORKER_TOKEN:-sched-worker-dev}
      POLL_TIMEOUT_SEC: ${POLL_TIMEOUT_SEC:-30}
      POLL_INTERVAL_MS: ${POLL_INTERVAL_MS:-1000}
      WORKER_TTL_SEC: ${WORKER_TTL_SEC:-90}
      RUNNING_LOCK_TTL_SEC: ${RUNNING_LOCK_TTL_SEC:-3600}
      INSTANCE_CLAIM_TIMEOUT_SEC: ${INSTANCE_CLAIM_TIMEOUT_SEC:-120}
      INSTANCE_SCAN_INTERVAL_SEC: ${INSTANCE_SCAN_INTERVAL_SEC:-30}
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:9000/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 20s

  backend:
    build:
      context: ..
      dockerfile: docker/Dockerfile
      args:
        DEBIAN_MIRROR: ${DEBIAN_MIRROR:-mirrors.aliyun.com}
    image: jarvis-core-backend:latest
    container_name: jarvis-core-backend
    restart: unless-stopped
    ports:
      - "${BACKEND_PORT:-666}:8000"
    env_file:
      - .env
    environment:
      SERVER_ADDR: ":8000"
      DB_PATH: /app/data/app.db
      UPLOAD_DIR: /app/data/uploads
      PUBLIC_BASE_URL: ${PUBLIC_BASE_URL:-http://localhost:666}
      SCHEDULER_SERVER_URL: ${SCHEDULER_SERVER_URL:-http://scheduler:9000}
      SCHEDULER_ENABLE: ${SCHEDULER_ENABLE:-true}
      SCHEDULER_ADMIN_TOKEN: ${SCHEDULER_ADMIN_TOKEN:-sched-admin-dev}
      SCHEDULER_WORKER_TOKEN: ${SCHEDULER_WORKER_TOKEN:-sched-worker-dev}
      SCHEDULER_POLL_TIMEOUT_SEC: ${SCHEDULER_POLL_TIMEOUT_SEC:-30}
      SCHEDULER_POLL_EMPTY_BACKOFF_MS: ${SCHEDULER_POLL_EMPTY_BACKOFF_MS:-200}
      SCHEDULER_POLL_IDLE_SEC: ${SCHEDULER_POLL_IDLE_SEC:-30}
    volumes:
      - backend-data:/app/data
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8000/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 25s
    depends_on:
      scheduler:
        condition: service_healthy

volumes:
  backend-data:
"""

SQL_PATCH = """-- 补全「系统管理」下菜单（旧库升级时使用）
-- mysql --default-character-set=utf8mb4 -uroot -p jarvis_core < patch_sys_menu_routes.sql

USE jarvis_core;
SET NAMES utf8mb4;

UPDATE sys_menu SET status = '1', hidden = 0, is_deleted = 0 WHERE id IN (1, 2, 3, 6, 7);

UPDATE sys_menu SET
  name = '菜单管理', title = '菜单管理',
  route_path = '/system/menu/index', component_path = 'system/menu/index',
  status = '1', hidden = 0, is_deleted = 0, updated_time = NOW()
WHERE id = 6;

UPDATE sys_menu SET
  name = '字典管理', title = '字典管理',
  route_path = '/system/dict/index', component_path = 'system/dict/index',
  status = '1', hidden = 0, is_deleted = 0, updated_time = NOW()
WHERE id = 7;

INSERT IGNORE INTO sys_role_menus (role_id, menu_id) VALUES
  (1, 1), (1, 2), (1, 3), (1, 6), (1, 7);
"""

FRONTEND_ENV_DEV = """VITE_APP_TITLE=Jarvis
VITE_BASE=/
VITE_API_BASE_URL=/api/v1
"""

FRONTEND_ENV_PROD = """VITE_APP_TITLE=Jarvis
VITE_BASE=/
VITE_API_BASE_URL=/api/v1
"""

SQL_REMOVE = [
    "backend/sql/seed_http_tool_is_arbitrate.sql",
    "backend/sql/seed_ai_eino_skill_call_yzx.sql",
    "backend/sql/patch_outbound_voice_menus.sql",
    "backend/sql/patch_ai_eino_tool_config_json.sql",
    "backend/sql/patch_ai_eino_skill.sql",
]


def write(rel: str, content: str) -> None:
    path = ROOT / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8", newline="\n")
    print("updated", rel)


def main() -> None:
    write("backend/.env.example", BACKEND_ENV)
    write("scheduler/.env.example", SCHEDULER_ENV)
    write("docker/.env.example", DOCKER_ENV)
    write("docker/docker-compose.yml", DOCKER_COMPOSE)
    write("backend/sql/patch_sys_menu_routes.sql", SQL_PATCH)
    write("frontend/web/.env.development", FRONTEND_ENV_DEV)
    write("frontend/web/.env.production", FRONTEND_ENV_PROD)

    for rel in SQL_REMOVE:
        path = ROOT / rel
        if path.exists():
            path.unlink()
            print("removed", rel)

    print("done")


if __name__ == "__main__":
    main()
