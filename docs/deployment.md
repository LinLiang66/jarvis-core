# 部署指南

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
