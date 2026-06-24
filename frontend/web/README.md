# jarvis-core 前端

Vue 3 + TypeScript + Element Plus 管理后台。

## 快速启动

```powershell
pnpm install
pnpm dev
```

默认 http://localhost:5050 ，API 代理至 `http://localhost:8000`。

## 脚本

| 命令 | 说明 |
|------|------|
| `pnpm dev` | 开发服务器 |
| `pnpm build` | 生产构建 |
| `pnpm preview` | 预览构建产物 |
| `pnpm typecheck` | TypeScript 检查 |
| `pnpm lint` | ESLint |

## 环境变量

| 变量 | 默认 |
|------|------|
| `VITE_APP_TITLE` | Jarvis |
| `VITE_BASE` | `/` |
| `VITE_API_BASE_URL` | `/api/v1` |

## 页面结构

```text
src/views/
├── dashboard/       # 工作台
├── system/          # 用户、角色、菜单、字典
├── openplatform/    # 应用、接口、文档、统计
└── login/           # 登录
```

更多见 [docs/architecture.md](../docs/architecture.md) 与 [docs/development.md](../docs/development.md)。
