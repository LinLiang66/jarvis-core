# 文档与配置维护示例

## 示例 1：新增「环境变量」说明

**场景**：后端新增 `JWT_REFRESH_DAYS`，需在模板和文档中体现。

1. 确认 `backend/internal/config/config.go` 已支持
2. 编辑 `scripts/update_docs_config.py` → `BACKEND_ENV` 增加一行
3. 编辑 `scripts/write_project_docs.py` → `GETTING_STARTED` 表格增加一行
4. 运行两个脚本：

```powershell
python scripts/update_docs_config.py
python scripts/write_project_docs.py
```

## 示例 2：只改 Docker 端口

**场景**：默认宿主机端口从 666 改为 8080。

1. 编辑 `scripts/update_docs_config.py`：
   - `DOCKER_ENV` 中 `BACKEND_PORT=8080`、`PUBLIC_BASE_URL=http://localhost:8080`
   - `DOCKER_COMPOSE` 中 `ports` 默认值（如有硬编码注释一并改）
2. 编辑 `scripts/write_project_docs.py`：
   - `DEPLOYMENT` 中端口说明
   - `DOCKER_README` 中访问地址
3. 运行：

```powershell
python scripts/update_docs_config.py
python scripts/write_project_docs.py
```

## 示例 3：新增 docs 章节

**场景**：增加 `docs/faq.md`。

1. 在 `write_project_docs.py` 顶部增加 `FAQ = """..."""` 
2. `main()` 增加 `write("docs/faq.md", FAQ)`
3. `DOCS_INDEX` 表格增加链接
4. 运行 `python scripts/write_project_docs.py`

## 示例 4：API 变更后同步文档

**场景**：新增 `GET /api/v1/system/dept/list`。

1. 改代码并合并
2. 编辑 `write_project_docs.py` 中 `API_REFERENCE` 增加章节
3. 若为新模块，同步 `ARCHITECTURE` 菜单表、`DEVELOPMENT` 扩展说明
4. 运行 `python scripts/write_project_docs.py`

## 反例

```powershell
# ❌ 直接改 docs/getting-started.md，未改脚本
# 下次 python scripts/write_project_docs.py 会覆盖

# ❌ 在 docker/.env.example 写 sk-xxxx 真实密钥

# ❌ 用 PowerShell Set-Content 批量替换中文 Markdown

# ❌ 只运行 update_docs_config 却期望 docs/ 更新
```
