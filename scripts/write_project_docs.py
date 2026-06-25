# -*- coding: utf-8 -*-
"""Write jarvis-core project documentation (UTF-8).

Deployment configs: scripts/update_docs_config.py
Agent skill: .cursor/skills/docs-config/SKILL.md
"""
from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"


def write(rel: str, content: str) -> None:
    path = ROOT / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content.strip() + "\n", encoding="utf-8", newline="\n")
    print("updated", rel)


DOCS_INDEX = """# 项目文档

jarvis-core 是一套前后端分离的管理后台基础框架，内置 **系统管理**（用户 / 角色 / 菜单 / 字典 / 存储配置 / 文件管理）与 **开放平台**（网关、应用、接口文档、调用统计），可作为业务系统的起点快速扩展。

## 文档目录

| 文档 | 说明 |
|------|------|
| [快速开始](getting-started.md) | 环境准备、本地启动、默认账号 |
| [架构说明](architecture.md) | 目录结构、模块划分、认证与权限 |
| [API 参考](api-reference.md) | 管理端 REST 接口一览 |
| [开放平台](openplatform.md) | 网关协议、握手流程、接入与扩展 |
| [部署指南](deployment.md) | Docker、生产环境、前后端分离部署 |
| [开发指南](development.md) | 二次开发、菜单扩展、新增 Action |

## 子项目说明

| 路径 | 文档 |
|------|------|
| 后端 | [backend/README.md](../backend/README.md) |
| 前端 | [frontend/web/README.md](../frontend/web/README.md) |
| Docker | [docker/README.md](../docker/README.md) |
| SDK 示例 | [examples/README.md](../examples/README.md) |

## 仓库与上游

| 远程 | 地址 | 用途 |
|------|------|------|
| origin | https://gitcode.com/LinLiang/jarvis-core.git | 开源基础框架（master） |
| upstream | https://github.com/lin-97/gi-element-plus-admin.git | 上游 UI 框架（main） |

同步上游 UI 框架：

```bash
git fetch upstream
git merge upstream/main
```
"""

GETTING_STARTED = """# 快速开始

## 环境要求

| 工具 | 版本 |
|------|------|
| Go | 1.21+（Docker 构建使用 1.23） |
| Node.js | 18+ |
| pnpm | 8+ |
| MySQL | 8.0+（生产推荐；本地可仅用 SQLite） |
| Redis | 6+（可选，登录 token 缓存） |
| Docker | 23+（仅容器部署时需要） |

## 1. 克隆仓库

```powershell
git clone https://gitcode.com/LinLiang/jarvis-core.git
cd jarvis-core
```

## 2. 启动后端

```powershell
cd backend
copy .env.example .env
go run ./cmd/server
```

- 默认监听 `:8000`
- 健康检查：`GET http://localhost:8000/health`
- 未配置 MySQL 时使用 SQLite：`backend/data/app.db`
- 首次启动自动建表并写入种子数据（含默认本地存储 `code=local`）

### 文件存储相关环境变量

在 `backend/.env` 中可配置（完整列表见 `backend/.env.example`）：

| 变量 | 说明 | 默认 |
|------|------|------|
| `UPLOAD_DIR` | 本地存储根目录（默认存储路径） | `./data/uploads` |
| `STATIC_URL_PREFIX` | 本地文件 HTTP 前缀 | `/static` |
| `PUBLIC_BASE_URL` | 对外访问基址（拼接文件 URL） | `http://127.0.0.1:8000` |
| `IMAGE_COMPRESS_ENABLE` | 上传图片是否智能压缩 | `true` |
| `IMAGE_COMPRESS_MAX_DIM` | 图片最长边上限（像素） | `1920` |
| `IMAGE_COMPRESS_QUALITY` | JPEG 压缩质量 1–100 | `85` |
| `IMAGE_COMPRESS_MIN_BYTES` | 小于该字节数不压缩 | `102400` |
| `IMAGE_COMPRESS_MAX_INPUT` | 参与压缩的单文件上限 | `20971520` |
| `DB_AUTO_MIGRATE` | 启动时是否 AutoMigrate（生产稳定库可设 `false`） | `true` |

对象存储（OSS）在管理端 **系统管理 → 存储配置** 中维护，支持 S3 兼容服务（阿里云 OSS、腾讯云 COS、MinIO 等）；可配置 **Base URL** 将返回链接域名替换为内网地址以节省公网流量。

### 启动性能

启动日志会输出 `[startup] ready in ...` 及各阶段耗时。优化策略：

- MySQL 与 Redis **并行连接**
- **单次** AutoMigrate（合并多表），避免重复迁移
- 增量菜单、默认存储种子 **后台异步** 执行
- 增量菜单补丁 **快速跳过**（路径已存在则不再全表扫描）
- Schema 补丁 **幂等检测**，避免每次 `ALTER TABLE`
- 生产环境可设 `DB_AUTO_MIGRATE=false` 跳过迁移以进一步加速

### 使用 MySQL（推荐）

在 `backend/.env` 中配置：

```env
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=your_password
MYSQL_DATABASE=jarvis_core
```

创建数据库：

```sql
CREATE DATABASE jarvis_core DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 3. 启动前端

新开终端：

```powershell
cd frontend/web
pnpm install
pnpm dev
```

- 默认访问：http://localhost:5050
- 开发环境通过 Vite 代理将 `/api`、`/static` 转发至 `http://localhost:8000`

## 4. 登录

| 字段 | 值 |
|------|-----|
| 用户名 | `admin` |
| 密码 | `123456` |

> 生产环境请立即修改默认密码与 `JWT_SECRET`。

## 5. 验证构建

```powershell
cd backend
go test ./...

cd ../frontend/web
pnpm build
```

## 常见问题

### 端口被占用

- 后端：修改 `backend/.env` 中 `SERVER_ADDR`，如 `:8080`
- 前端：修改 `frontend/web/vite.config.ts` 中 `server.port`，或设置 `VITE_API_PROXY_TARGET` 指向新后端地址

### Redis 未启动

默认 `REDIS_REQUIRED=false`，Redis 不可用时降级为纯 JWT，服务仍可启动。

### 旧库菜单缺失

若从旧版本升级 MySQL 库，可执行：

```powershell
mysql --default-character-set=utf8mb4 -uroot -p jarvis_core < backend/sql/patch_sys_menu_routes.sql
```
"""

ARCHITECTURE = """# 架构说明

## 总体架构

```text
┌─────────────┐     HTTP /api/v1      ┌──────────────────────────────┐
│  Vue 3 前端  │ ────────────────────► │  Go API (Gin)                │
│  :5050      │     Bearer JWT        │  :8000                       │
└─────────────┘                       │  ├─ auth（登录 / 用户信息）    │
                                      │  ├─ system（RBAC 管理）       │
                                      │  └─ openplatform（网关 + 管理）│
                                      └──────────┬───────────────────┘
                                                 │
                    ┌────────────────────────────┼────────────────────────────┐
                    ▼                            ▼                            ▼
              MySQL / SQLite                  Redis                    本地静态 / 对象存储
         （业务数据）              （登录 token + 开放平台会话 + 统计）   （/static 或 S3 兼容 OSS）
```

开放平台多副本部署时，Redis 用于共享 **token/3DES 会话**（键 `open:session:{token}`）及统计同步锁；详见 [开放平台 - 网关集群部署](openplatform.md#网关集群部署jarvis-服务端)。

## 项目目录

```text
jarvis-core/
├── backend/                 # Go REST API
│   ├── cmd/server/          # 入口
│   └── internal/
│       ├── config/          # 环境变量
│       ├── router/          # 路由装配
│       ├── handler/         # HTTP 层（auth、system、openplatform）
│       ├── middleware/      # JWT、超管
│       ├── service/
│       │   ├── openplatform/  # 网关、Action 注册、加解密
│       │   └── storage/       # 本地/OSS 上传、图片压缩、删除
│       ├── store/           # GORM 仓储
│       ├── model/           # 实体
│       └── database/        # 迁移与种子
├── frontend/web/            # Vue 3 管理后台
│   └── src/
│       ├── apis/            # 接口封装
│       ├── views/           # 页面（system、openplatform、dashboard）
│       ├── router/          # 路由与守卫
│       └── stores/          # Pinia 状态
├── docker/                  # 后端容器化
├── examples/                # 开放平台 SDK 示例（Go / Python / Java）
└── docs/                    # 项目文档（本目录）
```

## 后端分层

| 层 | 职责 |
|----|------|
| `handler` | 参数校验、HTTP 响应、调用 service/store |
| `service/openplatform` | 网关编排、Action 元数据、业务 handler 注册 |
| `store` | 数据库 CRUD |
| `middleware` | JWT 解析、登录态校验 |

路由注册见 `backend/internal/router/router.go`：

- 公开：`/health`、`/api/v1/auth/login`、`/api/v1/open/gateway`
- 鉴权：`/api/v1/system/*`、`/api/v1/open-app/*` 及 `/api/v1/auth/userinfo`

## 前端架构

- **动态路由**：登录后请求 `/api/v1/system/menu/routes`，按菜单树生成路由
- **权限**：路由 `meta` 与按钮级权限码控制可见性
- **UI**：Element Plus + gi-component（GiTable、FormDialog 等表格 CRUD 模式）
- **状态**：Pinia 管理用户、主题、标签页

## 认证流程

1. `POST /api/v1/auth/login` 提交用户名密码
2. 后端校验并签发 JWT；若 Redis 可用则缓存 token
3. 前端存储 token，后续请求携带 `Authorization: Bearer <token>`
4. `GET /api/v1/auth/userinfo` 获取用户信息与权限

## 内置菜单

| 模块 | 页面 |
|------|------|
| 工作台 | 仪表盘 |
| 系统管理 | 用户、角色、菜单、字典、存储配置、文件管理 |
| 开放平台 | 应用、接口、文档、统计 |

种子数据由 `backend/internal/database/seed_sys.go` 写入；默认创建本地存储 `local`（见 `seed_storage.go`）。

## 文件与存储

| 能力 | 说明 |
|------|------|
| 存储类型 | **本地存储**（磁盘目录 + `/static/{code}/` 访问）与 **对象存储**（S3 兼容 API） |
| 默认存储 | 首次启动自动创建 `code=local`；上传未指定 `storageId` 时使用默认存储 |
| OSS Base URL | 对象存储可填内网 `baseUrl`；上传返回 URL 会替换 OSS 域名为该地址（路径不变） |
| 图片压缩 | JPEG/PNG/WebP/BMP 上传时可选智能压缩（超尺寸缩放、JPEG 质量、仅当体积更小才替换） |
| 删除文件夹 | 递归删除子目录/文件的数据库记录，并删除本地文件或 OSS 对象，避免脏数据 |
| 静态访问 | 本地存储：`GET {PUBLIC_BASE_URL}{STATIC_URL_PREFIX}/{code}/...`；生产环境 Nginx 需代理 `/static/` |

存储引擎见 `backend/internal/service/storage/`；管理端页面：`views/system/storage`、`views/system/file`。
"""

API_REFERENCE = """# API 参考

> 管理端接口前缀：`/api/v1`  
> 除登录、健康检查、开放平台网关外，均需请求头：`Authorization: Bearer <token>`

统一响应格式（业务接口）：

```json
{
  "code": 200,
  "message": "success",
  "data": { }
}
```

## 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 服务存活探测 |

## 认证 `/api/v1/auth`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/login` | 登录，返回 token |
| GET | `/userinfo` | 当前用户信息（需登录） |
| POST | `/logout` | 退出（需登录） |

## 系统管理 `/api/v1/system`

### 用户 `/system/user`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/list` | 分页列表 |
| GET | `/:id` | 详情 |
| POST | `` | 新增 |
| PUT | `/:id` | 更新 |
| PUT | `/:id/password` | 重置密码 |
| PUT | `/:id/status` | 更新状态 |
| POST | `/delete` | 批量删除 |

### 角色 `/system/role`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/options` | 下拉选项 |
| GET | `/list` | 分页列表 |
| GET | `/:id` | 详情 |
| POST | `` | 新增 |
| PUT | `/:id` | 更新 |
| GET | `/:id/menus` | 获取角色菜单 |
| PUT | `/:id/menus` | 分配菜单 |
| POST | `/delete` | 批量删除 |

### 菜单 `/system/menu`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/routes` | 当前用户可访问路由（前端动态路由） |
| GET | `/tree` | 菜单树（管理页） |
| POST | `` | 新增 |
| PUT | `/:id` | 更新 |
| POST | `/delete` | 删除 |

### 字典 `/system/dict`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/type/list` | 字典类型列表 |
| POST | `/type` | 新增类型 |
| PUT | `/type/:id` | 更新类型 |
| POST | `/type/delete` | 删除类型 |
| GET | `/data/by-code/:code` | 按编码取字典项 |
| GET | `/data/list` | 字典项列表 |
| POST | `/data` | 新增字典项 |
| PUT | `/data/:id` | 更新字典项 |
| PUT | `/data/:id/status` | 更新状态 |
| POST | `/data/delete` | 删除字典项 |

### 存储配置 `/storage`

> 需超级管理员

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/list` | 列表（`type`：1 本地，2 对象存储） |
| GET | `/:id` | 详情 |
| POST | `` | 新增 |
| PUT | `/:id` | 更新 |
| PUT | `/:id/status` | 启用/禁用 |
| PUT | `/:id/default` | 设为默认存储 |
| POST | `/delete` | 批量删除 |

对象存储字段含 `accessKey`、`secretKey`、`endpoint`、`bucketName`、`baseUrl`（内网访问域名，可选）、`domain`（自定义域名，可选）。

### 文件管理 `/file`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/list` | 分页列表（`parentPath`、`storageId`、`originalName`） |
| GET | `/statistics` | 文件/目录数量与总大小 |
| POST | `/upload` | 上传（`multipart/form-data`：`file`、可选 `parentPath`、`storageId`） |
| POST | `/dir` | 创建文件夹（JSON：`parentPath`、`originalName`） |
| POST | `/delete` | 批量删除（超管；删文件夹会递归删子项及 OSS/本地对象） |

上传图片时若开启 `IMAGE_COMPRESS_*`，服务端自动压缩后再写入存储。

## 静态资源

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/static/{storageCode}/*` | 本地存储文件（无需 JWT） |

## 开放平台管理 `/api/v1/open-app`

> 需管理端登录

### 应用

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/list` | 应用列表 |
| GET | `/:id` | 应用详情 |
| POST | `` | 创建应用（生成 AppID / 密钥） |
| PUT | `/:id` | 更新 |
| POST | `/delete` | 删除 |
| POST | `/:id/regenerate-keys` | 重新生成密钥 |

### 接口 Action

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/action/list` | Action 列表 |
| GET | `/action/by-action` | 按 action 名查询 |
| POST | `/action/sync` | 从代码注册表同步 |
| POST | `/action` | 新增 |
| GET | `/action/:id` | 详情 |
| PUT | `/action/:id` | 更新 |
| POST | `/action/delete` | 删除 |

### 文档与统计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/doc` | 接口文档列表 |
| GET | `/doc/:action` | 单个 Action 文档 |
| GET | `/stat/daily` | 日统计 |
| GET | `/stat/logs` | 调用日志 |

## 开放平台网关

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/open/gateway` | 对外统一网关（`application/x-www-form-urlencoded`） |

网关协议、握手与加解密见 [开放平台](openplatform.md)。
"""

OPENPLATFORM = """# 开放平台

开放平台提供对外 **统一网关**，第三方应用通过 AppID、签名与可选 3DES 加密调用后端注册的 Action。

## 核心概念

| 概念 | 说明 |
|------|------|
| 应用（Open App） | 管理端创建，包含 AppID、SignSecret、AppSecret（RSA 私钥 DER Base64） |
| Action | 接口标识，如 `open.demo.echo` |
| 网关 | 单一入口 `POST /api/v1/open/gateway`，通过 `action` 参数路由 |
| 握手 Action | 获取 token、交换 3DES 密钥，不计费、不加密 |
| 会话（Session） | 一次握手产生 token + 3DES 密钥，业务请求必须携带同一 token |

## 内置 Action

| Action | 说明 | 加密 | 计费 |
|--------|------|------|------|
| `open.session.publickey` | 获取 token 与 RSA 公钥 | 否 | 否 |
| `microSession.create.secretkey` | 3DES 密钥交换 | 否 | 否 |
| `open.demo.echo` | Echo 演示，回显请求 JSON | 是 | 是 |

管理端 **开放平台 → 接口** 可查看文档；**应用** 页创建应用并授权 Action。

## 调用流程

```text
第三方客户端                          jarvis-core 网关
     │                                    │
     │  1. open.session.publickey         │
     │ ─────────────────────────────────► │
     │  ◄──────────── token + publicKey   │
     │                                    │
     │  2. microSession.create.secretkey  │
     │     (RSA 加密 clientPart)          │
     │ ─────────────────────────────────► │
     │  ◄──────────── serverPart          │
     │                                    │
     │  3. 业务 Action（3DES 加密 data）   │
     │ ─────────────────────────────────► │
     │  ◄──────────── 加密响应             │
```

## 会话与密钥

| 阶段 | 说明 |
|------|------|
| 获取 token | `publickey` 返回唯一 token，服务端创建会话 |
| 交换 3DES | `create.secretkey` 在同一 token 下写入 3DES 密钥（clientPart + serverPart） |
| 业务调用 | 携带 token + 3DES 加密 data；网关按 token 取密钥解密 |
| 续期 | 业务请求触发滑动续期（见下文服务端集群） |
| 失效 | token 过期或不存在时返回 **40001**，客户端需重新握手 |

> **重要**：同一 token 只能对应一套 3DES 密钥。若多个实例对**同一 token** 重复执行 `create.secretkey`，后写入的密钥会覆盖服务端记录，先完成握手的实例后续解密/调用将失败。

## 网关请求参数

公共字段（form-urlencoded）：

| 字段 | 说明 |
|------|------|
| `action` | Action 名称 |
| `appid` | 应用 AppID |
| `timestamp` / `req_time` | 毫秒时间戳（握手用 `timestamp`，业务用 `req_time`） |
| `version` | 固定 `V1.0` |
| `token` | 握手 step2 及业务 Action 必填 |
| `sign` | 签名 |
| `data` | 业务 JSON 字符串（握手阶段可为 `{}`） |

签名规则：对参与签名的参数按 key 排序拼接后，使用 SignSecret 做 MD5（详见 `examples/` SDK）。

错误码（网关 JSON `code`）：

| code | 含义 | 客户端处理 |
|------|------|------------|
| 200 | 成功 | — |
| 40001 | token 无效或过期 | 重新握手（集群场景见下文） |
| 40002 | 配额不足 | 检查应用配额或联系管理员 |

## 网关集群部署（jarvis-core 服务端）

jarvis-core API **多副本/多节点**部署时，**必须**配置共享 **Redis**，否则开放平台会话无法跨节点一致。

### 服务端行为

实现见 `backend/internal/service/openplatform/session_store.go`：

| 项 | 说明 |
|----|------|
| Redis 键 | `open:session:{token}` |
| 存储内容 | `app_id`、`token`、`tdes_key`、创建时间 |
| 会话 TTL | **2 小时** |
| 滑动续期 | 业务调用时，若剩余 TTL **< 30 分钟**，续期至 2 小时（任意节点执行即可全局生效） |
| 本地缓存 | 各节点懒加载 `TDESCipher`，减少重复构建；密钥以 Redis 为准 |

```text
                    ┌─────────┐     ┌─────────┐     ┌─────────┐
  客户端请求 ──────►│ API Pod │     │ API Pod │     │ API Pod │
                    └────┬────┘     └────┬────┘     └────┬────┘
                         │               │               │
                         └───────────────┼───────────────┘
                                         ▼
                               Redis（会话权威源）
                         open:session:{token} → SessionInfo
```

### 未启用 Redis 时

- 会话与 3DES 密钥仅存**当前进程内存**
- 负载均衡打到不同 Pod 时，极易出现 **token 找不到**、**3des key not initialized**、**40001**
- 仅适合本地单进程调试；**生产集群请设置 `REDIS_ENABLE=true` 并指向同一 Redis**

环境变量见 [部署指南](deployment.md) 与 `backend/.env.example` 中 `REDIS_*`。

## 调用方集群部署（业务系统 / SDK）

`examples/` 下 Go / Python / Java 示例为**单进程演示**，每次运行独立握手，**不包含** Redis 共享会话与分布式锁。

业务系统若以 **多实例、多 Pod、多容器** 调用 jarvis-core 开放平台，须自行实现 **集群版客户端**（可参考 `robot-dms-admin` 中 `continew-funcapi` 的 `OpenPlatformClient` 集群实现思路）。

### 为何需要 Redis + 加锁

| 问题 | 后果 |
|------|------|
| 各实例启动时各自握手 | 重复占用握手配额；若实例间误共享 token 又各自 `create.secretkey`，服务端 3DES 密钥被覆盖，部分节点调用失败 |
| 实例 A 已握手，实例 B 再次握手并覆盖共享缓存 | B 使用新 token，A 仍用旧 token → A 收到 **40001** 或解密失败 |
| token 过期后多实例同时重新握手 | 无锁时可能并发多次 `publickey` / `create.secretkey`，加剧上述竞态 |

目标：**同一 AppID 在整个集群内只维护一套「token + 3DES 密钥」**，且**全局仅一个实例**执行握手写入。

### 推荐架构

```text
  ┌──────────┐  ┌──────────┐  ┌──────────┐
  │ 业务 Pod │  │ 业务 Pod │  │ 业务 Pod │
  └────┬─────┘  └────┬─────┘  └────┬─────┘
       │             │             │
       │ ① 读共享会话 │             │
       │ ② 未命中则抢锁│             │
       └─────────────┼─────────────┘
                     ▼
              Redis（调用方独立实例）
    open:client:session:{appId}      → token + sessionKey
    open:client:session:lock:{appId} → 握手分布式锁
                     │
                     ▼ 仅持锁实例执行握手
              jarvis-core 开放平台网关
```

> 调用方 Redis 与 jarvis-core 服务端 Redis **相互独立**；键名前缀建议与业务隔离，避免与 `open:session:{token}`（服务端）混淆。

### Redis 数据设计（调用方）

| 键 | 示例 | 值 | 建议 TTL |
|----|------|-----|----------|
| 共享会话 | `open:client:session:{appId}` | JSON：`token`、`sessionKey`（3DES 合成密钥）、`createdAt` | **110 分钟**（略小于服务端 2h，到期前由客户端触发重新握手） |
| 握手锁 | `open:client:session:lock:{appId}` | 分布式锁 token | 锁持有 **30s**，等待 **60s** |

`sessionKey` 即客户端本地 `randomNum + serverPart`（与 `create.secretkey` 完成后一致），写入 Redis 后其他节点**直接复用**，**不要再调** `create.secretkey`。

### ensureSession 流程（调用方）

业务调用前统一走 `ensureSession()`（伪代码）：

```text
1. 本地已有 token + TDESCipher → 直接返回
2. 从 Redis 读取 open:client:session:{appId}
   → 命中则加载到本地，返回
3. 尝试获取 open:client:session:lock:{appId}
   → 未获取到：轮询 Redis 共享会话（如 30 次 × 1s），由持锁实例写入后加载
   → 获取到：再次 double-check Redis（防止重复握手）
4. 执行握手：publickey → create.secretkey
5. 将 {token, sessionKey} 写入 Redis（TTL 110min），释放锁
6. 初始化本地 TDESCipher
```

**主动重新握手**（如管理端轮换密钥、收到 40001）：

```text
1. 获取握手锁
2. 清除本地会话 + 删除 Redis 共享会话键
3. 执行完整握手并写回 Redis
4. 释放锁
```

业务请求遇到 **40001**：触发**集群级** `rehandshake()`（带锁），成功后**重试一次**原请求。

### 实现参数参考

以生产可用的集群客户端为参考（可按语言/中间件调整）：

| 参数 | 建议值 | 说明 |
|------|--------|------|
| 共享会话 TTL | 110 分钟 | 早于服务端 2h 过期，避免边缘时刻双方会话不一致 |
| 握手锁过期 | 30 秒 | 防止持锁进程崩溃导致死锁 |
| 抢锁最大等待 | 60 秒 | 与其他实例协调 |
| 等待共享会话 | 30 轮 × 1 秒 | 未抢到锁的实例等待持锁方写完 Redis |
| 进程内锁 | `synchronized` / mutex | 保护本实例内 token、cipher 字段并发读写 |

### 单实例 vs 集群

| 场景 | 要求 |
|------|------|
| 本地开发、CI、单次 demo | 可直接使用 `examples/`，每进程独立握手 |
| 生产多副本调用开放平台 | **必须** Redis 共享会话 + 握手分布式锁 |
| jarvis-core 网关多副本 | **必须** 服务端 Redis（见上一节） |

## 管理端操作

1. **创建应用**：开放平台 → 应用 → 新建，记录 AppID、SignSecret、AppSecret
2. **同步接口**：开放平台 → 接口 → 同步，将代码中注册的 Action 写入数据库
3. **授权应用**：为应用勾选可调用的 Action
4. **查看文档**：开放平台 → 文档
5. **监控调用**：开放平台 → 统计

## SDK 示例

`examples/` 提供 Go、Python、Java **单进程**客户端，演示握手与 Echo 调用（不含集群 Redis 逻辑）：

```powershell
cd examples/openplatform-go-demo
go run . -appid=app_xxx -sign=signSecret -secret=appSecretBase64
```

详见 [examples/README.md](../examples/README.md)。生产集群接入请按上文 **调用方集群部署** 扩展客户端。

## 扩展新 Action

1. 在 `backend/internal/service/openplatform/actions_meta.go` 注册元数据（标题、分类、是否加密/计费、请求响应 schema）
2. 在 `backend/internal/service/openplatform/business.go` 注册业务 handler
3. 重启后端，管理端执行 **接口同步**
4. 为应用授权新 Action

更详细的开发步骤见 [开发指南](development.md#扩展开放平台-action)。

## 集群部署检查清单

**jarvis-core 网关（服务端）**

- [ ] 多副本共用同一 Redis，`REDIS_ENABLE=true`
- [ ] 负载均衡健康检查包含 `GET /health`
- [ ] 开放平台统计同步依赖 Redis 分布式锁（已内置，需 Redis 可用）

**调用 jarvis-core 的业务系统（客户端）**

- [ ] 多实例共用 Redis，按 AppID 存储共享会话
- [ ] 握手路径有分布式锁，避免并发 `create.secretkey` 覆盖密钥
- [ ] 共享会话 TTL 小于服务端 2h，并实现 40001 集群重握手
- [ ] AppSecret、SignSecret 仅配置在服务端/密钥管理，勿硬编码在仓库
"""

DEPLOYMENT = """# 部署指南

## Docker 部署（仅后端）

当前 `docker/` 仅打包 **Go API**；MySQL、Redis 使用外部服务。

```powershell
cd docker
copy .env.example .env
# 编辑 .env：MYSQL_*、REDIS_*、JWT_SECRET、BACKEND_PORT
docker compose up -d --build
```

| 配置项 | 说明 | 默认 |
|--------|------|------|
| `BACKEND_PORT` | 宿主机映射端口 | `666` |
| `PUBLIC_BASE_URL` | 对外访问基址 | `http://localhost:666` |
| `JWT_SECRET` | 生产必改 | — |
| `MYSQL_*` | 生产推荐 MySQL | 空则容器内 SQLite |

验证：`GET http://localhost:666/health`

数据持久化：Docker volume `backend-data` → 容器 `/app/data`（SQLite 与上传文件）。

## 前端部署

### 构建

```powershell
cd frontend/web
pnpm install
pnpm build
```

产物目录：`frontend/web/dist`

### 环境变量（生产）

`.env.production`：

```env
VITE_APP_TITLE=Jarvis
VITE_BASE=/
VITE_API_BASE_URL=/api/v1
```

若 API 与前端同域，由 Nginx 反向代理 `/api` 至后端；若跨域，设置完整 API 地址并确保后端 CORS（开发环境已允许 `*`）。

### Nginx 示例

```nginx
server {
    listen 80;
    server_name admin.example.com;

    root /var/www/jarvis-core/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /static/ {
        proxy_pass http://127.0.0.1:8000;
    }
}
```

## 后端直接部署（无 Docker）

```powershell
cd backend
copy .env.example .env
# 配置 MYSQL_*、JWT_SECRET、REDIS_*
go build -o server ./cmd/server
./server
```

Linux 可使用 `run_linux.sh`；Windows 使用 `run_win.bat`（自动复制 `.env.example`）。

## 生产检查清单

- [ ] 修改默认管理员密码
- [ ] 设置强随机 `JWT_SECRET`
- [ ] 使用 MySQL 而非 SQLite
- [ ] **网关多副本**：Redis 启用且各 Pod 指向同一实例（开放平台会话必需，见 [openplatform.md](openplatform.md)）
- [ ] 配置 `PUBLIC_BASE_URL` 为真实域名（影响本地存储文件 URL 与默认存储访问路径）
- [ ] 使用对象存储时配置正确的 S3 Endpoint；内网访问可设 `baseUrl`
- [ ] Nginx 代理 `/static/` 至后端（本地文件访问）
- [ ] 前端 `VITE_BASE` 与 Nginx 路径一致
- [ ] 勿将 `backend/.env`、`docker/.env` 提交到版本库
- [ ] **业务侧多实例调用开放平台**：调用方实现 Redis 共享会话与握手分布式锁（见 [openplatform.md#调用方集群部署](openplatform.md#调用方集群部署业务系统--sdk)）
"""

DEVELOPMENT = """# 开发指南

## 技术栈约定

| 端 | 约定 |
|----|------|
| 后端 | Go 标准布局、`internal/` 不可外部导入、handler 薄 service/store 厚 |
| 前端 | Vue 3 Composition API + `<script setup>`、TypeScript、gi-component 表格页模式 |
| 提交 | Conventional Commits 中文说明，见 `.cursor/skills/git-commit/` |

## 本地开发流程

1. 后端 `go run ./cmd/server`（支持热重启需借助 air 等工具，项目未内置）
2. 前端 `pnpm dev`
3. 改 API 后运行 `go test ./...`
4. 改页面后运行 `pnpm typecheck` / `pnpm lint`

## 新增管理页面

1. **后端菜单**：在 `backend/internal/database/seed_sys.go` 或管理端「菜单管理」增加菜单项  
   - `route_path`：前端路由，如 `/biz/order/index`  
   - `component_path`：组件路径，如 `biz/order/index`（对应 `views/biz/order/index.vue`）
2. **前端页面**：在 `frontend/web/src/views/` 下新建 Vue 文件
3. **API 封装**：在 `frontend/web/src/apis/` 增加接口函数
4. **角色授权**：为角色分配新菜单

列表页可参考 `views/system/user/index.vue`（GiTable + FormDialog）；存储与文件管理见 `views/system/storage`、`views/system/file`。

## 新增 REST 模块

1. `internal/model` 定义实体
2. `internal/store` 实现仓储
3. `internal/handler` 实现 HTTP 处理器与 `Register`
4. `internal/router/router.go` 注册路由组
5. 在 `database/app.go` 挂载 store（若需新表）

## 扩展开放平台 Action

以 Echo 为例：

**1. 注册元数据** — `actions_meta.go`：

```go
RegisterActionMetaWithTypes(ActionMeta{
    Action: "your.biz.action",
    Title:  "业务接口",
    Category: "业务",
    Encrypted: true,
    Billable: true,
}, reqStruct{}, respStruct{})
```

**2. 注册 handler** — `business.go`：

```go
registerBusiness("your.biz.action", handleYourBiz)
```

**3. 同步**：管理端 → 开放平台 → 接口 → 同步

**4. 授权**：为测试应用勾选该 Action

**5. 验证**：使用 `examples/openplatform-go-demo` 或 Postman 调网关

**6. 集群调用**：若业务系统多实例部署，须按 [开放平台 - 调用方集群部署](openplatform.md#调用方集群部署业务系统--sdk) 实现 Redis 共享会话与握手锁；`examples/` 仅为单进程演示。

## 数据库补丁

- 新环境：依赖 GORM AutoMigrate + 种子，一般无需手工 SQL
- 旧库升级：使用 `backend/sql/` 下补丁，如 `patch_sys_menu_routes.sql`

## 从上游合并

前端 UI 能力来自 [gi-element-plus-admin](https://github.com/lin-97/gi-element-plus-admin)：

```bash
git fetch upstream
git merge upstream/main
```

合并后重点检查：`frontend/web/package.json`、路由守卫、gi-component 破坏性变更。

## 相关 Cursor Skills

| Skill | 用途 |
|-------|------|
| `table-page` | 生成标准 CRUD 列表页 |
| `openplatform-api` | 新增开放平台 Action 详细规范 |
| `git-commit` | 提交信息规范 |
"""

ROOT_README = """# jarvis-core

前后端分离的管理后台基础框架：**系统管理** + **开放平台**，适合作为业务系统的起点快速扩展。

## 特性

- RBAC 权限、动态菜单路由、JWT 登录
- 用户 / 角色 / 菜单 / 字典 / 存储配置 / 文件管理
- 本地存储 + S3 兼容对象存储、图片智能压缩、文件夹递归删除
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
"""

BACKEND_README = """# jarvis-core 后端

Go REST API：**Gin + GORM + JWT**，提供系统管理与开放平台能力。

## 快速启动

```powershell
cd backend
copy .env.example .env
go run ./cmd/server
```

健康检查：`GET http://localhost:8000/health`

## 模块

| 模块 | 路径前缀 |
|------|----------|
| 认证 | `/api/v1/auth` |
| 系统管理 | `/api/v1/user`、`/role`、`/menu`、`/dict`、`/storage`、`/file` 等 |
| 静态文件 | `/static/{storageCode}/` |
| 开放平台网关 | `/api/v1/open/gateway` |
| 开放平台管理 | `/api/v1/open-app/*` |

## 配置

复制 `.env.example` 为 `.env`。关键变量：

| 变量 | 说明 |
|------|------|
| `MYSQL_HOST` / `MYSQL_DATABASE` | 配置后使用 MySQL |
| `JWT_SECRET` | JWT 密钥（生产必改） |
| `REDIS_ENABLE` | 是否启用 Redis（**网关集群 + 开放平台会话强烈建议 true**） |
| `UPLOAD_DIR` / `PUBLIC_BASE_URL` | 本地存储目录与对外 URL 基址 |
| `IMAGE_COMPRESS_*` | 上传图片智能压缩（本地与 OSS 均生效） |

完整说明见 [docs/getting-started.md](../docs/getting-started.md) 与 [docs/api-reference.md](../docs/api-reference.md)。

## 测试

```powershell
go test ./...
```

## 部署

见 [docs/deployment.md](../docs/deployment.md) 与 [docker/README.md](../docker/README.md)。

模块名：`jarvis-core/backend`
"""

FRONTEND_README = """# jarvis-core 前端

Vue 3 + TypeScript + Element Plus 管理后台。

## 快速启动

```powershell
pnpm install
pnpm dev
```

默认 http://localhost:5050 ，API 代理至 `http://localhost:8000`。

## 脚本

| 命令 | 说明 |
|------|------|
| `pnpm dev` | 开发服务器 |
| `pnpm build` | 生产构建 |
| `pnpm preview` | 预览构建产物 |
| `pnpm typecheck` | TypeScript 检查 |
| `pnpm lint` | ESLint |

## 环境变量

| 变量 | 默认 |
|------|------|
| `VITE_APP_TITLE` | Jarvis |
| `VITE_BASE` | `/` |
| `VITE_API_BASE_URL` | `/api/v1` |

## 页面结构

```text
src/views/
├── dashboard/       # 工作台
├── system/          # 用户、角色、菜单、字典、存储、文件
├── openplatform/    # 应用、接口、文档、统计
└── login/           # 登录
```

更多见 [docs/architecture.md](../docs/architecture.md) 与 [docs/development.md](../docs/development.md)。
"""

DOCKER_README = """# Docker 部署

打包 **jarvis-core 后端 API** 容器；MySQL / Redis 需自行准备。

```powershell
cd docker
copy .env.example .env
docker compose up -d --build
```

默认 http://localhost:666/health

完整说明见 [docs/deployment.md](../docs/deployment.md)。
"""

EXAMPLES_README = """# 开放平台 SDK 示例

演示如何通过统一网关调用 jarvis-core 开放平台（握手、3DES 加密、Echo 演示接口）。

> 示例**仅调用内置演示 Action** `open.demo.echo`，不包含任何真实业务系统接口，避免泄露业务 Action 名称或参数结构。

## 目录

| 示例 | 语言 | 路径 |
|------|------|------|
| Go | Go 1.21+ | [openplatform-go-demo/](openplatform-go-demo/) |
| Python | Python 3.10+ | [openplatform-python-demo/](openplatform-python-demo/) |
| Java | Java 11+ | [openplatform-java-demo/](openplatform-java-demo/) |

## 前置条件

1. jarvis-core 后端已启动（默认 `http://127.0.0.1:8000`）
2. 管理端创建开放平台应用，记录 **AppID**、**SignSecret**、**AppSecret**（RSA 私钥 DER Base64）
3. 管理端 **接口 → 同步**，并为应用授权 **`open.demo.echo`**

## 调用流程

1. `open.session.publickey` — 获取 token
2. `microSession.create.secretkey` — 3DES 密钥交换
3. `open.demo.echo` — 加密 Echo 演示（回显请求 JSON）

## Go

```powershell
cd openplatform-go-demo
go run . -appid=app_xxx -sign=signSecret -secret=appSecretBase64
```

环境变量：`OPEN_GATEWAY_URL`、`OPEN_APP_ID`、`OPEN_SIGN_SECRET`、`OPEN_APP_SECRET`

## Python

```powershell
cd openplatform-python-demo
pip install -r requirements.txt
python demo.py --appid app_xxx --sign signSecret --secret appSecretBase64
```

## Java

```powershell
cd openplatform-java-demo
mvn -q exec:java -Dexec.args="http://127.0.0.1:8000/api/v1/open/gateway app_xxx signSecret appSecretBase64"
```

协议细节见 [docs/openplatform.md](../docs/openplatform.md)。
"""


def main() -> None:
    write("docs/README.md", DOCS_INDEX)
    write("docs/getting-started.md", GETTING_STARTED)
    write("docs/architecture.md", ARCHITECTURE)
    write("docs/api-reference.md", API_REFERENCE)
    write("docs/openplatform.md", OPENPLATFORM)
    write("docs/deployment.md", DEPLOYMENT)
    write("docs/development.md", DEVELOPMENT)
    write("README.md", ROOT_README)
    write("backend/README.md", BACKEND_README)
    write("frontend/web/README.md", FRONTEND_README)
    write("docker/README.md", DOCKER_README)
    write("examples/README.md", EXAMPLES_README)
    print("done")


if __name__ == "__main__":
    main()
