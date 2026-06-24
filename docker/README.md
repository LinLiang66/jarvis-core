# Docker 部署

打包 **jarvis 后端 API** 容器；MySQL / Redis 需自行准备。

```powershell
cd docker
copy .env.example .env
docker compose up -d --build
```

默认 http://localhost:666/health

完整说明见 [docs/deployment.md](../docs/deployment.md)。
