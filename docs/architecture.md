# 架构说明

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
