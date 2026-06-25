# jarvis-core 后端

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
