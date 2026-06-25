# 快速开始

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
