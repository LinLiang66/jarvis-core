# 开放平台

开放平台提供对外 **统一网关**，第三方应用通过 AppID、签名与可选 3DES 加密调用后端注册的 Action。

## 核心概念

| 概念 | 说明 |
|------|------|
| 应用（Open App） | 管理端创建，包含 AppID、SignSecret、AppSecret（RSA 私钥 DER Base64） |
| Action | 接口标识，如 `open.demo.echo` |
| 网关 | 单一入口 `POST /api/v1/open/gateway`，通过 `action` 参数路由 |
| 握手 Action | 获取 token、交换 3DES 密钥，不计费、不加密 |
| 会话（Session） | 一次握手产生 token + 3DES 密钥，业务请求必须携带同一 token |

## 内置 Action

| Action | 说明 | 加密 | 计费 |
|--------|------|------|------|
| `open.session.publickey` | 获取 token 与 RSA 公钥 | 否 | 否 |
| `microSession.create.secretkey` | 3DES 密钥交换 | 否 | 否 |
| `open.demo.echo` | Echo 演示，回显请求 JSON | 是 | 是 |

管理端 **开放平台 → 接口** 可查看文档；**应用** 页创建应用并授权 Action。

## 调用流程

```text
第三方客户端                          jarvis-core 网关
     │                                    │
     │  1. open.session.publickey         │
     │ ─────────────────────────────────► │
     │  ◄──────────── token + publicKey   │
     │                                    │
     │  2. microSession.create.secretkey  │
     │     (RSA 加密 clientPart)          │
     │ ─────────────────────────────────► │
     │  ◄──────────── serverPart          │
     │                                    │
     │  3. 业务 Action（3DES 加密 data）   │
     │ ─────────────────────────────────► │
     │  ◄──────────── 加密响应             │
```

## 会话与密钥

| 阶段 | 说明 |
|------|------|
| 获取 token | `publickey` 返回唯一 token，服务端创建会话 |
| 交换 3DES | `create.secretkey` 在同一 token 下写入 3DES 密钥（clientPart + serverPart） |
| 业务调用 | 携带 token + 3DES 加密 data；网关按 token 取密钥解密 |
| 续期 | 业务请求触发滑动续期（见下文服务端集群） |
| 失效 | token 过期或不存在时返回 **40001**，客户端需重新握手 |

> **重要**：同一 token 只能对应一套 3DES 密钥。若多个实例对**同一 token** 重复执行 `create.secretkey`，后写入的密钥会覆盖服务端记录，先完成握手的实例后续解密/调用将失败。

## 网关请求参数

公共字段（form-urlencoded）：

| 字段 | 说明 |
|------|------|
| `action` | Action 名称 |
| `appid` | 应用 AppID |
| `timestamp` / `req_time` | 毫秒时间戳（握手用 `timestamp`，业务用 `req_time`） |
| `version` | 固定 `V1.0` |
| `token` | 握手 step2 及业务 Action 必填 |
| `sign` | 签名 |
| `data` | 业务 JSON 字符串（握手阶段可为 `{}`） |

签名规则：对参与签名的参数按 key 排序拼接后，使用 SignSecret 做 MD5（详见 `examples/` SDK）。

错误码（网关 JSON `code`）：

| code | 含义 | 客户端处理 |
|------|------|------------|
| 200 | 成功 | — |
| 40001 | token 无效或过期 | 重新握手（集群场景见下文） |
| 40002 | 配额不足 | 检查应用配额或联系管理员 |

## 网关集群部署（jarvis-core 服务端）

jarvis-core API **多副本/多节点**部署时，**必须**配置共享 **Redis**，否则开放平台会话无法跨节点一致。

### 服务端行为

实现见 `backend/internal/service/openplatform/session_store.go`：

| 项 | 说明 |
|----|------|
| Redis 键 | `open:session:{token}` |
| 存储内容 | `app_id`、`token`、`tdes_key`、创建时间 |
| 会话 TTL | **2 小时** |
| 滑动续期 | 业务调用时，若剩余 TTL **< 30 分钟**，续期至 2 小时（任意节点执行即可全局生效） |
| 本地缓存 | 各节点懒加载 `TDESCipher`，减少重复构建；密钥以 Redis 为准 |

```text
                    ┌─────────┐     ┌─────────┐     ┌─────────┐
  客户端请求 ──────►│ API Pod │     │ API Pod │     │ API Pod │
                    └────┬────┘     └────┬────┘     └────┬────┘
                         │               │               │
                         └───────────────┼───────────────┘
                                         ▼
                               Redis（会话权威源）
                         open:session:{token} → SessionInfo
```

### 未启用 Redis 时

- 会话与 3DES 密钥仅存**当前进程内存**
- 负载均衡打到不同 Pod 时，极易出现 **token 找不到**、**3des key not initialized**、**40001**
- 仅适合本地单进程调试；**生产集群请设置 `REDIS_ENABLE=true` 并指向同一 Redis**

环境变量见 [部署指南](deployment.md) 与 `backend/.env.example` 中 `REDIS_*`。

## 调用方集群部署（业务系统 / SDK）

`examples/` 下 Go / Python / Java 示例为**单进程演示**，每次运行独立握手，**不包含** Redis 共享会话与分布式锁。

业务系统若以 **多实例、多 Pod、多容器** 调用 jarvis-core 开放平台，须自行实现 **集群版客户端**（可参考 `robot-dms-admin` 中 `continew-funcapi` 的 `OpenPlatformClient` 集群实现思路）。

### 为何需要 Redis + 加锁

| 问题 | 后果 |
|------|------|
| 各实例启动时各自握手 | 重复占用握手配额；若实例间误共享 token 又各自 `create.secretkey`，服务端 3DES 密钥被覆盖，部分节点调用失败 |
| 实例 A 已握手，实例 B 再次握手并覆盖共享缓存 | B 使用新 token，A 仍用旧 token → A 收到 **40001** 或解密失败 |
| token 过期后多实例同时重新握手 | 无锁时可能并发多次 `publickey` / `create.secretkey`，加剧上述竞态 |

目标：**同一 AppID 在整个集群内只维护一套「token + 3DES 密钥」**，且**全局仅一个实例**执行握手写入。

### 推荐架构

```text
  ┌──────────┐  ┌──────────┐  ┌──────────┐
  │ 业务 Pod │  │ 业务 Pod │  │ 业务 Pod │
  └────┬─────┘  └────┬─────┘  └────┬─────┘
       │             │             │
       │ ① 读共享会话 │             │
       │ ② 未命中则抢锁│             │
       └─────────────┼─────────────┘
                     ▼
              Redis（调用方独立实例）
    open:client:session:{appId}      → token + sessionKey
    open:client:session:lock:{appId} → 握手分布式锁
                     │
                     ▼ 仅持锁实例执行握手
              jarvis-core 开放平台网关
```

> 调用方 Redis 与 jarvis-core 服务端 Redis **相互独立**；键名前缀建议与业务隔离，避免与 `open:session:{token}`（服务端）混淆。

### Redis 数据设计（调用方）

| 键 | 示例 | 值 | 建议 TTL |
|----|------|-----|----------|
| 共享会话 | `open:client:session:{appId}` | JSON：`token`、`sessionKey`（3DES 合成密钥）、`createdAt` | **110 分钟**（略小于服务端 2h，到期前由客户端触发重新握手） |
| 握手锁 | `open:client:session:lock:{appId}` | 分布式锁 token | 锁持有 **30s**，等待 **60s** |

`sessionKey` 即客户端本地 `randomNum + serverPart`（与 `create.secretkey` 完成后一致），写入 Redis 后其他节点**直接复用**，**不要再调** `create.secretkey`。

### ensureSession 流程（调用方）

业务调用前统一走 `ensureSession()`（伪代码）：

```text
1. 本地已有 token + TDESCipher → 直接返回
2. 从 Redis 读取 open:client:session:{appId}
   → 命中则加载到本地，返回
3. 尝试获取 open:client:session:lock:{appId}
   → 未获取到：轮询 Redis 共享会话（如 30 次 × 1s），由持锁实例写入后加载
   → 获取到：再次 double-check Redis（防止重复握手）
4. 执行握手：publickey → create.secretkey
5. 将 {token, sessionKey} 写入 Redis（TTL 110min），释放锁
6. 初始化本地 TDESCipher
```

**主动重新握手**（如管理端轮换密钥、收到 40001）：

```text
1. 获取握手锁
2. 清除本地会话 + 删除 Redis 共享会话键
3. 执行完整握手并写回 Redis
4. 释放锁
```

业务请求遇到 **40001**：触发**集群级** `rehandshake()`（带锁），成功后**重试一次**原请求。

### 实现参数参考

以生产可用的集群客户端为参考（可按语言/中间件调整）：

| 参数 | 建议值 | 说明 |
|------|--------|------|
| 共享会话 TTL | 110 分钟 | 早于服务端 2h 过期，避免边缘时刻双方会话不一致 |
| 握手锁过期 | 30 秒 | 防止持锁进程崩溃导致死锁 |
| 抢锁最大等待 | 60 秒 | 与其他实例协调 |
| 等待共享会话 | 30 轮 × 1 秒 | 未抢到锁的实例等待持锁方写完 Redis |
| 进程内锁 | `synchronized` / mutex | 保护本实例内 token、cipher 字段并发读写 |

### 单实例 vs 集群

| 场景 | 要求 |
|------|------|
| 本地开发、CI、单次 demo | 可直接使用 `examples/`，每进程独立握手 |
| 生产多副本调用开放平台 | **必须** Redis 共享会话 + 握手分布式锁 |
| jarvis-core 网关多副本 | **必须** 服务端 Redis（见上一节） |

## 管理端操作

1. **创建应用**：开放平台 → 应用 → 新建，记录 AppID、SignSecret、AppSecret
2. **同步接口**：开放平台 → 接口 → 同步，将代码中注册的 Action 写入数据库
3. **授权应用**：为应用勾选可调用的 Action
4. **查看文档**：开放平台 → 文档
5. **监控调用**：开放平台 → 统计

## SDK 示例

`examples/` 提供 Go、Python、Java **单进程**客户端，演示握手与 Echo 调用（不含集群 Redis 逻辑）：

```powershell
cd examples/openplatform-go-demo
go run . -appid=app_xxx -sign=signSecret -secret=appSecretBase64
```

详见 [examples/README.md](../examples/README.md)。生产集群接入请按上文 **调用方集群部署** 扩展客户端。

## 扩展新 Action

1. 在 `backend/internal/service/openplatform/actions_meta.go` 注册元数据（标题、分类、是否加密/计费、请求响应 schema）
2. 在 `backend/internal/service/openplatform/business.go` 注册业务 handler
3. 重启后端，管理端执行 **接口同步**
4. 为应用授权新 Action

更详细的开发步骤见 [开发指南](development.md#扩展开放平台-action)。

## 集群部署检查清单

**jarvis-core 网关（服务端）**

- [ ] 多副本共用同一 Redis，`REDIS_ENABLE=true`
- [ ] 负载均衡健康检查包含 `GET /health`
- [ ] 开放平台统计同步依赖 Redis 分布式锁（已内置，需 Redis 可用）

**调用 jarvis-core 的业务系统（客户端）**

- [ ] 多实例共用 Redis，按 AppID 存储共享会话
- [ ] 握手路径有分布式锁，避免并发 `create.secretkey` 覆盖密钥
- [ ] 共享会话 TTL 小于服务端 2h，并实现 40001 集群重握手
- [ ] AppSecret、SignSecret 仅配置在服务端/密钥管理，勿硬编码在仓库
