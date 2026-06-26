# 任务调度

jarvis-core 的任务调度采用 **scheduler-server + jarvis Worker + 前端 BFF** 三层架构。

## 架构概览

```text
┌──────────────┐   /api/v1/scheduler/*    ┌─────────────────┐
│  Vue 3 前端   │ ───────────────────────► │  jarvis 后端     │
│  管理任务实例  │   JWT + BFF 代理         │  :8000           │
└──────────────┘                          │  schedulerclient │
                                          └────────┬─────────┘
                                                   │ /worker/v1/*
                                                   ▼
                                          ┌─────────────────┐
                                          │ scheduler-server │
                                          │ :9000            │
                                          │ /admin/v1/*      │
                                          └────────┬─────────┘
                     ┌─────────────────────────────┼─────────────────────────┐
                     ▼                             ▼                         ▼
              MySQL jarvis_scheduler            Redis                   Cron / 实例扫描
              （任务定义、实例、日志）           （队列、心跳、锁）
```

| 角色 | 说明 |
|------|------|
| **scheduler-server** | 独立 Go 服务，负责任务 CRUD、Cron 触发、实例分发与状态机 |
| **jarvis Worker** | 内嵌于 backend 的 `schedulerclient`，向 scheduler 注册 handler 并长轮询执行 |
| **jarvis BFF** | 将前端管理请求代理至 scheduler `/admin/v1/*`，注入 `X-Scheduler-Token` |

**部署建议**：**1 个 scheduler-server** + **N 个 jarvis 后端**（水平扩展 Worker 能力）。scheduler 建议单实例；多个 backend 共享同一 scheduler 与 Redis。

## 端口与数据库

| 项 | 值 |
|----|-----|
| scheduler-server | `:9000`，健康检查 `GET /health` |
| jarvis 后端 | `:8000`（Docker 宿主机默认 `:666`） |
| scheduler MySQL 库 | **`jarvis_scheduler`**（与后端 `jarvis_core` 分离） |
| Redis | scheduler **必需**（任务队列、Worker 心跳、运行锁） |

## 本地启动

### 1. 准备 MySQL 与 Redis

```sql
CREATE DATABASE jarvis_scheduler DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 启动 scheduler-server

```powershell
cd scheduler
copy .env.example .env
go run ./cmd/server
```

配置见 `scheduler/.env.example`：MySQL、Redis、`ADMIN_TOKEN`、`WORKER_TOKEN` 及引擎参数（`POLL_*`、`INSTANCE_*` 等）。

### 3. 启用 jarvis Worker + BFF

```powershell
cd backend
copy .env.example .env
```

在 `.env` 中设置：

```env
SCHEDULER_ENABLE=true
SCHEDULER_SERVER_URL=http://127.0.0.1:9000
SCHEDULER_ADMIN_TOKEN=sched-admin-dev
SCHEDULER_WORKER_TOKEN=sched-worker-dev
```

`SCHEDULER_ADMIN_TOKEN` / `SCHEDULER_WORKER_TOKEN` 必须与 scheduler 的 `ADMIN_TOKEN` / `WORKER_TOKEN` 一致。

启动后端后，Worker 自动注册内置 handler（如 `demo.hello`、`stat.sync`），并开始长轮询。

## Docker Compose

```powershell
cd docker
copy .env.example .env
docker compose up -d --build
```

compose 包含 `scheduler` 与 `backend` 服务；`backend` 依赖 `scheduler` 健康检查通过后再启动。环境变量模板见 `docker/.env.example`。

## 环境变量

### scheduler-server（`scheduler/.env`）

| 变量 | 说明 | 默认 |
|------|------|------|
| `SERVER_ADDR` | 监听地址 | `:9000` |
| `MYSQL_*` / `MYSQL_DSN` | MySQL 连接，库名 `jarvis_scheduler` | 见 `.env.example` |
| `REDIS_*` | Redis（`REDIS_REQUIRED=true`） | `127.0.0.1:6379` |
| `ADMIN_TOKEN` | 管理 API 鉴权（BFF 使用） | `sched-admin-dev` |
| `WORKER_TOKEN` | Worker API 鉴权 | `sched-worker-dev` |
| `POLL_TIMEOUT_SEC` | Worker 长轮询超时（秒） | `30` |
| `POLL_INTERVAL_MS` | 引擎扫描间隔（毫秒） | `1000` |
| `WORKER_TTL_SEC` | Worker 心跳过期（秒） | `90` |
| `RUNNING_LOCK_TTL_SEC` | 实例运行锁 TTL（秒） | `3600` |
| `INSTANCE_CLAIM_TIMEOUT_SEC` | 实例认领超时（秒） | `120` |
| `INSTANCE_SCAN_INTERVAL_SEC` | 实例扫描间隔（秒） | `30` |

### jarvis 后端（`backend/.env`）

| 变量 | 说明 | 默认 |
|------|------|------|
| `SCHEDULER_ENABLE` | 是否启用 Worker + BFF | `false` |
| `SCHEDULER_SERVER_URL` | scheduler 地址 | `http://127.0.0.1:9000` |
| `SCHEDULER_ADMIN_TOKEN` | BFF 代理鉴权 | 与 `ADMIN_TOKEN` 一致 |
| `SCHEDULER_WORKER_TOKEN` | Worker 客户端鉴权 | 与 `WORKER_TOKEN` 一致 |
| `SCHEDULER_INSTANCE_ID` | Worker 实例 ID（空则自动生成 UUID） | — |
| `SCHEDULER_POLL_TIMEOUT_SEC` | 长轮询超时，建议与 scheduler 一致 | `30` |
| `SCHEDULER_POLL_EMPTY_BACKOFF_MS` | 空 poll 退避 | `200` |
| `SCHEDULER_POLL_IDLE_SEC` | 调度门关闭时空闲重检间隔 | `30` |

Docker 部署时上述变量在 `docker/.env.example` 中以 `SCHEDULER_*` 前缀统一维护。

## API 路径

| 受众 | 前缀 | 鉴权 |
|------|------|------|
| 管理端（经 BFF） | `/api/v1/scheduler/*` → scheduler `/admin/v1/*` | JWT + `X-Scheduler-Token` |
| Worker | scheduler `/worker/v1/*` | `X-Scheduler-Token`（Worker token） |
| 健康检查 | `GET /health` | 无 |

管理 API 示例：`GET /admin/v1/jobs`、`POST /admin/v1/jobs/:id/trigger`、`GET /admin/v1/instances`。

## 扩展 Worker Handler

在 `backend/internal/schedulerclient/` 注册 handler 名称（须与任务定义中的 `handler` 字段一致），参考 `handlers.go` 中 `RegisterDemoHelloHandler`。启动时 `StartDefault` 会注册内置 handler 并启动轮询循环。

## 相关代码

| 路径 | 说明 |
|------|------|
| `scheduler/cmd/server/main.go` | scheduler 入口 |
| `scheduler/internal/service/engine.go` | 调度引擎 |
| `backend/internal/schedulerclient/` | Worker 客户端 |
| `backend/internal/handler/scheduler/` | BFF 反向代理 |
