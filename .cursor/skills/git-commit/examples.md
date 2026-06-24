# Git 提交示例

## 示例 1：系统管理前端

**变更**：`frontend/web/src/views/system/user/**`

```powershell
git add frontend/web/src/views/system/user/ frontend/web/src/apis/system/user.ts
git commit -m "feat: 用户管理列表与表单" -m "GiTable 列表、FormDialog 与角色分配"
```

## 示例 2：开放平台前后端联调

**变更**：`backend/internal/service/openplatform/` + `frontend/web/src/views/openplatform/`

```powershell
git add backend/internal/service/openplatform/ backend/internal/handler/openplatform/ frontend/web/src/views/openplatform/
git commit -m "feat: 开放平台应用管理与网关接口" -m "后端 Action 注册与前端应用 CRUD"
```

## 示例 3：仅后端 Go

**变更**：`backend/internal/router/`、`backend/cmd/server/`

```powershell
git add backend/internal/router/router.go backend/cmd/server/main.go
git commit -m "feat(backend): 注册系统管理与开放平台路由"
```

## 示例 4：文档与配置

```powershell
git add README.md backend/README.md docker/.env.example
git commit -m "docs: 更新 jarvis 部署说明与环境变量模板"
```

## 示例 5：修复 bug

```powershell
git add frontend/web/src/router/guard.ts frontend/web/src/router/route-load-state.ts
git commit -m "fix(frontend): 登录后动态路由重复加载" -m "使用 route-load-state 标记避免二次 addRoute"
```

## 示例 6：数据库补丁

```powershell
git add backend/sql/patch_sys_menu_routes.sql
git commit -m "chore(backend): 系统管理菜单路由补丁 SQL"
```

## 反例（不要这样做）

```powershell
# ❌ 未获用户同意就提交
git add -A; git commit -m "update"

# ❌ 提交敏感文件
git add backend/.env docker/.env

# ❌ 提交 Go 编译产物
git add backend/server backend/*.exe

# ❌ 提交构建缓存
git add frontend/web/node_modules/

# ❌ 失败后 amend
git commit --amend  # hook 刚失败时禁止
```

## 提交说明对照

| 变更性质 | 推荐 type | 示例 subject |
|----------|-----------|--------------|
| 新列表页 | feat | feat: 字典管理列表页 |
| 修接口 404 | fix | fix(backend): 菜单树接口路径 |
| 抽 openplatform 模块 | refactor | refactor(backend): 拆分 gateway 与 app 服务 |
| 仅改 SCSS | style | style: 工作台卡片样式 |
| 加 skill | chore | chore: 新增 git-commit skill |
| SQL 补丁 | chore | chore(backend): 菜单路由 seed 补丁 |
