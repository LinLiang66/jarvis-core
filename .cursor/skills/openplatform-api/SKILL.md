---
name: openplatform-api
description: >-
  根据第三方接口地址、请求体、响应体 JSON，自动生成 funcapi 包（constants/types/agent/方法）
  并从 URL 推导 action/常量/函数名，注册开放平台（service.go、actions_meta.go、handler_*.go）。
  适用于新增开放平台接口、对接第三方 HTTP API、扩展韵达/funcapi 代理等场景。
---

# 开放平台第三方 API 接入

参考实现：`backend/internal/pkg/funcapi/yunda/`、`backend/internal/service/openplatform/handler_yunda.go`。

## 用户需提供

向用户确认或从对话中提取以下信息（缺一则先追问）：

| 字段 | 说明 | 示例 |
|------|------|------|
| 接口地址 | 完整 URL（含 HTTP 方法） | `GET https://customer-problem.yundasys.com/yrh-xt/app/common/getDictItems` |
| 请求体 JSON | 开放平台 `data` 解密后的业务 JSON（即 funcapi 入参） | `{"orgCode":"xxx","userId":"1",...}` |
| 响应体 JSON | 第三方原始响应或 `data` 字段结构 | `{"success":true,"code":200,"data":[...]}` |

**以下由 Agent 从 URL 自动推导，生成前向用户展示推导结果供确认；用户可覆盖：**

| 自动推导 | 说明 |
|----------|------|
| action | 开放平台 action，格式 `{provider}.{domain}.{method}` |
| Action 常量 | `Action{Provider}{Method}` |
| 包名 / provider | funcapi 子包名 |
| 业务分类 | 开放平台文档 Category |
| endpoint 常量 | `{method}Endpoint` |
| Go 函数名 | PascalCase(method) |
| Req/Resp 命名 | `{Method}Req` / 响应 struct |

可选（用户补充）：接口标题、默认请求头、multipart、必填校验、分页默认值、action 覆盖。

## action 自动生成（Step 0，必做）

**禁止让用户手写 action。** 接入前先解析 URL，按下列规则生成命名族，并在回复中列出推导表。

### 0.1 解析 URL

```
GET https://customer-problem.yundasys.com/yrh-xt/app/common/getDictItems
     └─ host ─────────────────────┘ └──────── path ──────────────────┘
```

- `apiBaseURL` = `scheme://host`（无尾斜杠）
- `endpoint` = path（以 `/` 开头，含 query 前的完整 path）
- `httpMethod` = GET / POST / …

### 0.2 host → provider + domain + category

查表匹配 host（子域优先）；未命中则用 host 首段作 provider，domain 默认 `api`，category 默认 `{provider} 接口`。

| host | provider | domain | category（文档分类） |
|------|----------|--------|----------------------|
| `customer-problem.yundasys.com` | `yunda` | `problem` | `韵达问题件` |
| `tmmweb.yundasys.com:4406` | `tmmweb` | `tmm` | `韵达TMM` |
| `yx.yundasys.com` | `yx` | `platform` | `韵达全链路` |

> 新系统接入时在 SKILL 本表或 `funcapi/{provider}/constants.go` 注释中追加一行映射。

### 0.3 path → method（action 第三段）

1. path 按 `/` 分段，**丢弃**噪声段：`yrh-xt`、`app`、`common`、`api`、`v1`、`v2`、空串
2. 取剩余段 `... , prev, last`（至少 1 段）

**规则 A — last 已是动词短语**（匹配 `^(get|set|create|update|delete|query|search|release|cancel)[A-Z]\w*$`）  
→ `method = last`（保持 camelCase）  
例：`getDictItems` → `getDictItems`，`getReceivingSite` → `getReceivingSite`

**规则 B — last 为泛化动作词**（`list` | `upload` | `query` | `search` | `detail` | `save` | `delete` | `update`）  
→ 与 `prev` 组合：

| last | 组合规则 | 示例 |
|------|----------|------|
| `list` | `get` + shorten(prev) + `List` | `problemManagementReport/list` → `getProblemManageList` |
| `upload` | 见 0.4 路径覆盖表；否则 `upload` + PascalCase(prev) | `file/upload` → `uploadFile` |
| 其他 | `{last}` + PascalCase(prev) | `query` + `Order` → `queryOrder` |

**shorten(resource)**：camelCase 后依次去掉尾部 `Report`、`Management`、`Info`（各最多一次）。  
例：`problemManagementReport` → `problemManage`

**规则 C — 其余**  
→ `method = toCamelCase(last)`；若仅一段且含 `-`，取最后有意义片段（如 `upload-img` → 看下一段）

### 0.4 路径覆盖表（优先于规则 B）

语义与 path 不一致时，**本表优先**（与仓库已有 action 对齐）：

| path 后缀（endpoint 包含） | method | 说明 |
|---------------------------|--------|------|
| `/upload-img/file/upload` | `uploadImage` | 问题件图片上传 |
| `/formData/upload` | `releaseIssue` | 发布问题件（非字面 upload） |
| `/expectedClosure/summary` | `getExpectedClosureSummary` | 预期结案汇总（path 为 resource/summary，对齐分页报表 get* 命名） |
| `/busi_intercept_benchmark/addData` | `addBusiInterceptBenchmarkData` | 拦截基准录入（避免 Rule C 仅得 addData） |
| `/config/qllgj/reply/pageData` | `getReplyPageData` | 全链路回复配置分页（避免 Rule C 仅得 pageData） |

新接口若语义特殊，接入后在本表追加一行。

### 0.5 组装 action 与关联命名

```
action = "{provider}.{domain}.{method}"
```

| 推导项 | 规则 | 示例（getDictItems） |
|--------|------|----------------------|
| action | 上式 | `yunda.problem.getDictItems` |
| Action 常量 | `Action` + PascalCase(provider) + PascalCase(method) | `ActionYundaGetDictItems` |
| endpoint 常量 | camelCase(method) + `Endpoint` | `getDictItemsEndpoint` |
| Go 函数 | PascalCase(method) | `GetDictItems` |
| 请求 struct | PascalCase(method) + `Req` | `GetDictItemsReq` |
| validate 函数 | `validate` + PascalCase(method) | `validateGetDictItems` |
| handler 函数 | `handle` + Provider + PascalCase(method)（特殊 handler 才需要） | `handleYundaUploadImage` |

**PascalCase(method)**：`getDictItems` → `GetDictItems`（首字母大写，其余保持）。

### 0.6 推导示例

| URL path | method | action |
|----------|--------|--------|
| `/yrh-xt/app/common/getDictItems` | `getDictItems` | `yunda.problem.getDictItems` |
| `/upload-img/file/upload` | `uploadImage` | `yunda.problem.uploadImage` |
| `/yrh-xt/app/formData/upload` | `releaseIssue` | `yunda.problem.releaseIssue` |
| `/yrh-xt/app/common/getReceivingSite` | `getReceivingSite` | `yunda.problem.getReceivingSite` |
| `/yrh-xt/app/problemManagementReport/list` | `getProblemManageList` | `yunda.problem.getProblemManageList` |
| `/yrh-xt/app/expectedClosure/summary` | `getExpectedClosureSummary` | `yunda.problem.getExpectedClosureSummary` |
| `/yrh-xt/app/problemManagementReport/autoClosureRules` | `autoClosureRules` | `yunda.problem.autoClosureRules` |
| `/yrh-xt/app/report/check/queryPage` | `queryPage` | `yunda.problem.queryPage` |
| `/platform-api/fullChainOrder/list` | `getFullChainOrderList` | `yx.platform.getFullChainOrderList` |
| `/platform-api/config/qllgj/reply/pageData` | `getReplyPageData` | `yx.platform.getReplyPageData` |
| `/platform-api/client/attachment/uploadForFullLink` | `uploadForFullLink` | `yx.platform.uploadForFullLink` |
| `/platform-api/expense/queryExpenseList` | `queryExpenseList` | `yx.platform.queryExpenseList` |
| `/platform-api/orderEnum/getAppealTypeEnum` | `getAppealTypeEnum` | `yx.platform.getAppealTypeEnum` |

生成代码前输出推导表；若 path 命中覆盖表或规则 A/B 有歧义，一句话说明选用理由。

## 文件结构（funcapi 子包）

每个第三方系统一个子包，**实体与实现分离**：

```
backend/internal/pkg/funcapi/{provider}/
├── constants.go   # apiBaseURL + endpoint 路径常量
├── types.go       # 请求/响应 struct（*Req 须嵌入 AgentRequestBase；json tag 与第三方一致）
├── agent.go       # agentGET/agentPOST、defaultHeaders（复用或新建）
└── {domain}.go    # 对外 API 函数（如 problem.go）
```

开放平台注册（同 provider 共用一个 handler 文件）：

```
backend/internal/service/openplatform/
├── service.go          # Action 常量
├── actions_meta.go     # 文档元数据（Title/Category/Description/Sort）
└── handler_{provider}.go  # init 注册 + validate + 特殊 handler
```

## 命名规范

action 及关联命名**一律走 Step 0 自动生成**；JSON → Go struct 规则：

- 字段保持 camelCase，`json` tag 与第三方一致
- 未知类型用 `interface{}`；数组元素单独建 struct
- 响应外层 `{success,code,message,data}` 用 `Response[T]`，T 为 `data` 类型
- **每个 `*Req` 第一个字段必须嵌入 `funcapi.AgentRequestBase`**（无业务字段时也保留，用于携带 `header`）

## AgentRequestBase（必做）

开放平台 `data` 可携带自定义 HTTP 头，由 `funcapi.PrepareAgentRequest` 从请求体提取后传给 IPC Agent，**不会**转发给第三方 JSON body。

```go
// funcapi/agent_request.go
type AgentRequestBase struct {
    Header map[string]string `json:"header,omitempty"`
}
```

调用方示例：

```json
{
  "header": { "Token": "custom-token", "App": "synergy" },
  "orgCode": "123456"
}
```

**强制规则：**

1. `types.go` 中**所有** `*Req` struct 第一行嵌入 `funcapi.AgentRequestBase`（含无参查询如 `GetAppealTypeEnumReq`）
2. `agent.go` 的 `agentGET` / `agentPOST` 必须调用 `PrepareAgentRequest` + `MergeAgentHeaders`（见「新建 provider 包」模板）
3. multipart 上传在 `{domain}.go` 内先 `PrepareAgentRequest`，再 `DoAgentMultipartRequest`
4. **禁止**从 `*Req` 移除嵌入、**禁止** agent 直连序列化 body 而不提取 header

## 实施步骤

复制 checklist 并逐项完成：

```
- [ ] Step 0: 从 URL 推导 action / Action 常量 / 函数名 / endpoint 常量（展示推导表）
- [ ] Step 1: 解析输入，确定 provider 包是否存在
- [ ] Step 2: constants.go 追加 endpoint（新包则含 apiBaseURL）
- [ ] Step 3: types.go 追加 Req/Resp struct（**每个 *Req 嵌入 funcapi.AgentRequestBase**）
- [ ] Step 4: {domain}.go 实现 API 函数
- [ ] Step 5: service.go 追加 Action 常量
- [ ] Step 6: actions_meta.go 注册文档元数据
- [ ] Step 7: handler_{provider}.go 注册业务 handler
- [ ] Step 8: go build ./... 验证编译
```

### Step 1：判断传输方式

| 场景 | funcapi 调用 |
|------|-------------|
| JSON GET | `agentGET[T](path, req, headers)` |
| JSON POST | `agentPOST[T](path, req, headers)` |
| multipart 文件 | `funcapi.DoAgentMultipartRequest(url, "POST", parts, &resp, headers)` |
| 开放平台侧 multipart/base64 | 在 handler 中校验后调 funcapi（见 UploadImage） |

第三方统一响应 `{success,code,message,data}` 时，在 types.go 定义：

```go
type Response[T any] struct {
    Success bool   `json:"success"`
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}
```

函数返回 `(Response[T], error)`，T 填 `data` 的类型。

### Step 2：constants.go

```go
const (
    apiBaseURL           = "https://example.com"
    getXxxEndpoint       = "/path/to/api"
)
```

已有包只追加 endpoint；URL 与现有 base 不同则评估是否新建 provider 包。

### Step 3：types.go

从用户 JSON 生成 struct，加中文注释说明用途。**每个 `*Req` 第一字段嵌入 `funcapi.AgentRequestBase`：**

```go
import "jarvis-core/backend/internal/pkg/funcapi"

// GetXxxReq 查询 xxx 请求。
type GetXxxReq struct {
    funcapi.AgentRequestBase
    OrgCode string `json:"orgCode"`
    UserId  string `json:"userId"`
}

// GetAppealTypeEnumReq 无业务参数时也须嵌入，仅用于携带 header。
type GetAppealTypeEnumReq struct {
    funcapi.AgentRequestBase
}
```

响应 struct（如 `XxxItem`、分页 `data` 内层）**不需要**嵌入 `AgentRequestBase`。

### Step 4：API 函数模板

**GET/POST（标准 JSON）**

```go
// GetXxx 查询 xxx。
func GetXxx(req GetXxxReq) (Response[[]XxxItem], error) {
    // 可选：默认值
    if req.PageNum <= 0 {
        req.PageNum = 1
    }
    return agentGET[[]XxxItem](getXxxEndpoint, req, nil)
    // POST 则 agentPOST[...](...)
}
```

**multipart 上传**（须 `PrepareAgentRequest` 提取 header，再合并默认头）

```go
func UploadXxx(req UploadXxxReq) (Response[UploadXxxResponse], error) {
    cleanBody, reqHeaders := funcapi.PrepareAgentRequest(req)
    req = cleanBody.(UploadXxxReq)
    contentType := req.ContentType
    if contentType == "" {
        contentType = "application/octet-stream"
    }
    headers := funcapi.MergeAgentHeaders(defaultProblemHeaders, reqHeaders)
    var resp Response[UploadXxxResponse]
    err := funcapi.DoAgentMultipartRequest(
        apiBaseURL+uploadXxxEndpoint, "POST",
        []funcapi.MultipartPart{{
            Name: "file", Type: "file",
            Filename: req.Filename, ContentType: contentType, Data: req.Data,
        }},
        &resp, headers,
    )
    return resp, err
}
```

`UploadXxxReq` 同样须在 `types.go` 嵌入 `funcapi.AgentRequestBase`。

### Step 5：service.go 追加常量

```go
const (
    // ...
    ActionYundaGetXxx = "yunda.problem.getXxx"
)
```

### Step 6：actions_meta.go

在 `registerBuiltinActionMetas()` 中追加（Sort 同分类递增）：

```go
RegisterActionMeta(ActionMeta{
    Action: ActionYundaGetXxx, Title: "查询 xxx", Category: "韵达问题件",
    Description: "一句话说明。",
    Encrypted: true, Billable: true, Sort: 150,
})
```

韵达类接口：`Encrypted: true, Billable: true`。握手/演示接口参考已有 `ActionEcho`。

Schema 无需手写：`registerYundaHandler` 会从实体自动生成 RequestSchema/ResponseSchema。

### Step 7：handler 注册

**标准 JSON 接口** — 在 `handler_{provider}.go` 的 `init()` 中：

```go
registerYundaHandler(ActionYundaGetXxx, validateGetXxx, yunda.GetXxx)
```

追加校验函数：

```go
func validateGetXxx(req yunda.GetXxxReq) error {
    return requireOrgUser(req.OrgCode, req.UserId)
    // 或: if strings.TrimSpace(req.OrderNo) == "" { return errors.New("orderNo required") }
}
```

`registerYundaHandler` 已封装：JSON 解析 → validate → 调 funcapi → `{action,data}` 响应 → 自动 `ApplyActionTypeSchemas`。

**特殊接口（如 base64 上传）** — 不用 registerYundaHandler，手动注册：

```go
ApplyActionTypeSchemas(ActionYundaUploadImage, yunda.UploadImageReq{}, yunda.Response[yunda.UploadImageResponse]{})
registerBusiness(ActionYundaUploadImage, handleYundaUploadImage)
```

handler 内完成 trim、base64 校验、大小限制等，最后 `marshalYundaResponse(action, data)`。

### Step 8：验证

```powershell
cd backend; go build ./...
```

重启服务后，开放平台文档页应出现新接口（启动时 `RegistryToModels` 同步 MySQL）。

## 新建 provider 包（首次接入）

1. 创建 `funcapi/{provider}/` 四文件：`constants.go`、`types.go`、`agent.go`、`{domain}.go`
2. `agent.go` 从 yunda 复制并改包名（**必须**含 `PrepareAgentRequest`，不可直连 body）：

```go
func agentGET[T any](path string, body any, headers map[string]string) (Response[T], error) {
    cleanBody, reqHeaders := funcapi.PrepareAgentRequest(body)
    headers = funcapi.MergeAgentHeaders(headers, reqHeaders)
    var resp Response[T]
    err := funcapi.DoAgentRequest(apiBaseURL+path, "GET", cleanBody, &resp, headers)
    return resp, err
}

func agentPOST[T any](path string, body any, headers map[string]string) (Response[T], error) {
    cleanBody, reqHeaders := funcapi.PrepareAgentRequest(body)
    headers = funcapi.MergeAgentHeaders(headers, reqHeaders)
    var resp Response[T]
    err := funcapi.DoAgentRequest(apiBaseURL+path, "POST", cleanBody, &resp, headers)
    return resp, err
}
```

3. `types.go`：**每个** `*Req` 第一字段嵌入 `funcapi.AgentRequestBase`（见上文「AgentRequestBase（必做）」）。
4. 新建 `handler_{provider}.go`，可复制 `handler_yunda.go` 骨架，将 `registerYundaHandler` 泛型注册函数**留在该文件**或抽到 `handler_provider.go` 共用（同 package 内）。
5. 若新 provider 的注册模式与韵达不同，可新建 `register{Provider}Handler` 但保持相同流程：EnsureActionMeta → ApplyActionTypeSchemas → registerBusiness。

## 禁止事项

- ❌ 绕过 `@/internal/pkg/funcapi` 的 `DoAgentRequest` / `DoAgentMultipartRequest` 直接 HTTP
- ❌ `*Req` 不嵌入 `funcapi.AgentRequestBase`，或 agent 不调用 `PrepareAgentRequest`（会导致 `data.header` 丢失）
- ❌ 实体与 API 函数混在同一文件（应 types.go + {domain}.go）
- ❌ 只加 funcapi 不注册 openplatform（文档与网关无法调用）
- ❌ 只注册 action 不写 funcapi 实现
- ❌ 在用户未要求时改 demo 示例或写 README

## 完整示例

见 [examples.md](examples.md)（韵达 getDictItems 端到端样例）。
