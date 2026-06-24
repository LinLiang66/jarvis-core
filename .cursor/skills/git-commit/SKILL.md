---
name: git-commit
description: >-
  按 jarvis-core 仓库规范将代码提交并可选推送到远程。涵盖变更审查、暂存范围、Conventional 风格
  中文提交说明、Git 安全约束与 Windows/PowerShell 命令。适用于用户说提交代码、git commit、
  推送到仓库、帮我 commit、保存到 git 等场景。
---

# Git 提交代码

## 前置约束（必须遵守）

1. **仅在用户明确要求时**才执行 `git add` / `git commit` / `git push`；未明确要求时只展示建议的提交说明与待提交文件列表。
2. **禁止**修改 git 配置（`git config`）。
3. **禁止**破坏性命令：`push --force`、`reset --hard` 等，除非用户在本轮对话中明确要求。
4. **禁止**跳过钩子：`--no-verify`、`--no-gpg-sign`，除非用户明确要求。
5. **禁止**向 `main`/`master` 强推；若用户要求强推，先警告风险。
6. **禁止**在提交说明中添加任何署名行（如 `Co-authored-by`、`Signed-off-by`、`Made-with: Cursor` 等）；`-m` 正文只写 type/scope 与中文说明。
7. **禁止** `git commit --amend`，除非同时满足：
   - 用户明确要求 amend，或 pre-commit 成功但自动改动了文件需纳入本次提交
   - HEAD 为本会话内创建且未 push
8. 提交失败或被 hook 拒绝时：**修复后新建 commit**，不要 amend 失败的提交。
9. 用**简体中文**向用户汇报进度与结果。

> **关于 Cursor 自动追加 `Co-authored-by`**：Agent 通过终端执行 `git commit` 时，Cursor IDE 可能在 hook/包装层自动注入 `Co-authored-by: Cursor <cursoragent@cursor.com>`，并非 skill 正文写入。若需纯净提交说明，请在 Cursor **Settings → Agents → Attribution**（或 Git 相关项）关闭 commit attribution，或在本地终端自行 `git commit`。

## 仓库结构速查

| 目录 | 说明 |
|------|------|
| `frontend/web/` | Vue 3 前端（路径别名 `@/` → `src/`） |
| `backend/` | Go 后端（`cmd/server`、`internal/`） |
| `backend/sql/` | 数据库补丁与种子 SQL |
| `.cursor/` | Cursor rules / skills |

提交前以 `.gitignore` 为准；常见**不应提交**：

- `.env`、`backend/.env`、`frontend/web/.env`
- `node_modules`、`backend/bin/`、`backend/*.exe`、`backend/server`
- `backend/logs`、`backend/data/`、`backend/*.db`
- 构建产物 `dist`、`frontend/dist`

## 标准工作流

### Step 1：并行收集状态（一次消息内执行）

```powershell
git status
git diff
git diff --staged
git log -5 --oneline
```

若用户只要 push 已有提交，可跳过 diff，执行 `git status` 与 `git log -3 --oneline`。

### Step 2：审查并划分提交范围

- 只暂存与本次任务相关的文件；不要整仓 `git add .` 除非用户明确要求或变更确实全部相关。
- 发现密钥、`.env`、编译产物误入时：**警告用户并排除**，不要提交。
- 前后端联动功能可同一 commit；无关重构应拆分为多次提交（用户要求单次提交时除外）。

### Step 3：撰写提交说明

遵循本仓库既有风格（见 `git log`）：

```
<type>[optional scope]: <简短中文描述>

[可选：1-2 句说明动机或影响范围]
```

**type 常用值**：

| type | 用途 |
|------|------|
| `feat` | 新功能 |
| `fix` | 缺陷修复 |
| `refactor` | 重构（无行为变化） |
| `style` | 样式/UI，不改逻辑 |
| `chore` | 构建、依赖、配置、skills/rules |
| `docs` | 文档（用户明确要求时） |

**scope 可选**（与历史提交一致）：`frontend`、`backend`，或省略。

示例：

```
feat: AI 工具管理与 HTTP Tool 配置页

新增工具 CRUD、配置编辑器与联调测试弹窗
```

```
fix(frontend): 修复动态路由重复加载
```

```
feat(backend): 开放平台网关 Action 注册与文档
```

说明聚焦 **why**，避免罗列文件名。

### Step 4：暂存并提交（顺序执行）

```powershell
git add <path1> <path2> ...
git commit -m "<subject>" -m "<body>"
```

**Windows PowerShell** 多行说明用多个 `-m`（不要依赖 bash HEREDOC）：

```powershell
git commit -m "feat: 开放平台应用管理页" -m "应用 CRUD、密钥展示与接口授权"
```

**Git Bash** 可用 HEREDOC：

```bash
git commit -m "$(cat <<'EOF'
feat: AI 对话页与 A2UI 渲染

流式 reasoning、工具调用与 Markdown 展示
EOF
)"
```

### Step 5：验证

```powershell
git status
```

确认 working tree 干净或仅剩用户知悉的未提交文件。

### Step 6：推送（仅用户明确要求时）

```powershell
git push
# 新分支首次：
git push -u origin HEAD
```

推送前确认当前分支与远程跟踪关系；若落后远程，先说明是否需要 `git pull --rebase`，不要擅自 force。

## 推送 + PR（用户要求发 PR 时）

1. 并行：`git status`、`git diff`、`git branch -vv`、`git log origin/main..HEAD --oneline`（基分支按仓库默认，常见 `main`/`master`）。
2. 无提交则先按上文完成 commit。
3. `git push -u origin HEAD`（需网络权限）。
4. 使用 `gh pr create` 创建 PR（需 `gh` 已登录）。

## 完成后汇报模板

向用户说明：

- 提交 hash（短 hash 即可）
- 提交说明全文
- 包含的主要路径
- 是否已 push
- 仍留在工作区的未提交文件（如有）

## 故障处理

| 情况 | 处理 |
|------|------|
| pre-commit / lint 失败 | 修复问题 → **新 commit**，不 amend |
| 无变更可提交 | 告知用户，不创建空 commit |
| 冲突 | 不自动 force；指导 pull/rebase 或让用户决定 |
| 误暂存敏感文件 | `git reset HEAD <file>` 后从暂存区移除 |

## 更多示例

见 [examples.md](examples.md)
