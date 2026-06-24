# 开发指南

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

列表页可参考现有 `views/system/user/index.vue`（GiTable + FormDialog）。

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
