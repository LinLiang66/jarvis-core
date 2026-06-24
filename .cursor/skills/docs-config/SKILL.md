---
name: docs-config
description: >-
  维护 jarvis-core 项目文档与部署配置。通过 scripts/write_project_docs.py 更新 Markdown 文档，
  通过 scripts/update_docs_config.py 更新 .env.example、docker-compose 等配置模板。
  适用于用户要求改 README、docs/、部署说明、环境变量模板、Docker 配置，或提到 write_project_docs、
  update_docs_config、项目文档、部署配置等场景。
---

# jarvis-core 文档与部署配置维护

## 核心原则

1. **单一数据源**：内容只改脚本里的常量/字符串，**不要**直接编辑生成后的 Markdown 或 `.env.example`（下次跑脚本会被覆盖）。
2. **两个脚本分工明确**，禁止混用：
   - 文档 → `scripts/write_project_docs.py`
   - 部署配置 → `scripts/update_docs_config.py`
3. **中文 Markdown 必须用 Python 脚本写入**（`encoding="utf-8"`），避免 Windows 下 Write 工具或 PowerShell 破坏编码。
4. **`.env` 注释用英文**，减少终端/编辑器乱码；文档正文用中文。
5. **禁止**在模板中提交真实密钥、密码、API Key；`docker/.env`、`backend/.env` 已在 `.gitignore` 中。

## 脚本分工

| 脚本 | 用途 | 生成/更新文件 |
|------|------|----------------|
| `scripts/write_project_docs.py` | 项目文档 | `docs/*.md`、`README.md`、`backend/README.md`、`frontend/web/README.md`、`docker/README.md`、`examples/README.md` |
| `scripts/update_docs_config.py` | 部署配置模板 | `backend/.env.example`、`docker/.env.example`、`docker/docker-compose.yml`、`frontend/web/.env.development`、`frontend/web/.env.production`、`backend/sql/patch_sys_menu_routes.sql` |

## 决策：改哪个脚本？

```
用户要改什么？
├── README / docs/ / API 说明 / 架构 / 开放平台文档 / 开发指南
│   └── 编辑 write_project_docs.py → 运行 write_project_docs.py
├── 环境变量 / Docker 端口 / compose / 前端 VITE_* / SQL 补丁
│   └── 编辑 update_docs_config.py → 运行 update_docs_config.py
└── 代码行为变更（新增 API、新 env 变量被 Go 读取）
    ├── 先改 backend/internal/config 等代码
    ├── 同步 update_docs_config.py 中的 BACKEND_ENV / DOCKER_ENV
    └── 若需对外说明，再改 write_project_docs.py 对应章节
```

## 工作流：更新文档

1. 打开 `scripts/write_project_docs.py`，找到对应常量（如 `GETTING_STARTED`、`API_REFERENCE`）。
2. 修改内容；新增文档时在 `main()` 增加 `write("docs/xxx.md", XXX)` 并在 `DOCS_INDEX` 加链接。
3. 若部署说明与 `docs/deployment.md` 重复，**只维护 `DEPLOYMENT` 常量**，子目录 README 保持简短并链到 docs。
4. 运行：

```powershell
python scripts/write_project_docs.py
```

5. 用 Read 工具抽查 `docs/README.md` 或根 `README.md` 中文是否正常（勿仅信终端输出编码）。

## 工作流：更新部署配置

1. 打开 `scripts/update_docs_config.py`，修改 `BACKEND_ENV`、`DOCKER_ENV`、`DOCKER_COMPOSE`、`FRONTEND_ENV_*` 等常量。
2. 新增环境变量时：
   - 确认 `backend/internal/config/config.go` 已读取该变量
   - 在 `BACKEND_ENV` 与 `DOCKER_ENV` 同步添加（Docker 可省略仅本地开发的项）
   - 在 `write_project_docs.py` 的 `GETTING_STARTED` 或 `DEPLOYMENT` 补充说明（若用户可见）
3. 运行：

```powershell
python scripts/update_docs_config.py
```

4. 提醒用户：本地 `docker/.env` 需自行 `copy .env.example .env` 后合并已有值，脚本**不会**写 `docker/.env`。

## 文档结构速查

```text
docs/
├── README.md           # 索引
├── getting-started.md  # 快速开始
├── architecture.md     # 架构
├── api-reference.md    # API
├── openplatform.md     # 开放平台
├── deployment.md       # 部署（与 docker/README 呼应）
└── development.md      # 二次开发
```

## 配置模板速查

| 常量 | 输出文件 |
|------|----------|
| `BACKEND_ENV` | `backend/.env.example` |
| `DOCKER_ENV` | `docker/.env.example` |
| `DOCKER_COMPOSE` | `docker/docker-compose.yml` |
| `FRONTEND_ENV_DEV` / `FRONTEND_ENV_PROD` | `frontend/web/.env.development` / `.production` |
| `SQL_PATCH` | `backend/sql/patch_sys_menu_routes.sql` |

## 禁止事项

- ❌ 用 Cursor Write 直接改含大量中文的 `docs/*.md`（易乱码）
- ❌ 在 `update_docs_config.py` 里维护 README（已移除，避免覆盖 `write_project_docs.py`）
- ❌ 在 `.env.example` 写真实密钥或已删除业务的变量（LLM、DASHSCOPE、LICENSE 等）
- ❌ 提交 `backend/.env`、`docker/.env`
- ❌ 只改生成文件不改脚本（下次运行会丢失）

## 变更后建议验证

文档/配置类改动通常不需要全量测试；若涉及端口或 API 路径描述，可选：

```powershell
cd backend && go test ./...
cd frontend/web && pnpm build
```

## 提交说明参考

```
docs: 更新开放平台接入说明
docs: 补充 Docker 部署检查清单
chore: 同步 backend/.env.example 新增 REDIS_* 变量
```

详见 `.cursor/skills/git-commit/`。

## 更多示例

见 [examples.md](examples.md)。
