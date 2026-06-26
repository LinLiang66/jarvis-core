# Docker 部署

部署 **scheduler-server**（`:9000`）与 **jarvis 后端 API**（容器 `:8000`，宿主机默认 `:666`）；MySQL、Redis 使用外部服务。

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

- **拓扑**：1 scheduler + N jarvis 后端 Worker（compose 默认各 1 副本）
- scheduler 独立库 **`jarvis_scheduler`**，与 `jarvis_core` 分离

完整说明见 [docs/deployment.md](../docs/deployment.md) 与 [docs/scheduler.md](../docs/scheduler.md)。
