# 示例：接入 getDictItems

> **通用约定**：下文所有 `*Req` 均须嵌入 `funcapi.AgentRequestBase` 作为第一字段；multipart 接口在函数内先 `PrepareAgentRequest` 再发请求。详见 [SKILL.md](SKILL.md)「AgentRequestBase（必做）」。

## 用户输入

```
接口: GET https://customer-problem.yundasys.com/yrh-xt/app/common/getDictItems

请求体:
{
  "orgCode": "123456",
  "orgName": "测试网点",
  "userId": "10001",
  "userName": "张三",
  "dictType": "order_type_code",
  "includeExt": 1
}

响应体:
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": [
    { "itemValue": "1", "itemText": "破损", "extField1": "" }
  ]
}
```

## Step 0：action 自动推导

| 项 | 推导过程 | 结果 |
|----|----------|------|
| host | `customer-problem.yundasys.com` | provider=`yunda`, domain=`problem` |
| path 去噪 | 去掉 `yrh-xt`,`app`,`common` → 剩 `getDictItems` | — |
| method | 规则 A：已是 `get*` 动词短语 | `getDictItems` |
| action | `yunda.problem.getDictItems` | |
| Action 常量 | `ActionYundaGetDictItems` | |
| endpoint 常量 | `getDictItemsEndpoint` | |
| Go 函数 | `GetDictItems` | |
| apiBaseURL | `https://customer-problem.yundasys.com` | |
| endpoint | `/yrh-xt/app/common/getDictItems` | |

## 生成结果

### constants.go（追加）

```go
getDictItemsEndpoint = "/yrh-xt/app/common/getDictItems"
```

### types.go（追加）

```go
import "jarvis/backend/internal/pkg/funcapi"

type DictItem struct {
    ItemValue string `json:"itemValue"`
    ItemText  string `json:"itemText"`
    ExtField1 string `json:"extField1,omitempty"`
    // ...
}

// GetDictItemsReq 问题件类型字典查询请求。
type GetDictItemsReq struct {
    funcapi.AgentRequestBase
    OrgCode      string `json:"orgCode"`
    OrgName      string `json:"orgName"`
    UserId       string `json:"userId"`
    UserName     string `json:"userName"`
    DictType     string `json:"dictType,omitempty"`
    AddAllAtHead int    `json:"addAllAtHead,omitempty"`
    IncludeExt   int    `json:"includeExt,omitempty"`
}
```

开放平台调用方可选传入 `header` 覆盖默认 Token/App：

```json
{
  "header": { "Token": "custom-token" },
  "orgCode": "123456",
  "userId": "10001",
  "dictType": "order_type_code"
}
```

### problem.go（追加）

```go
func GetDictItems(req GetDictItemsReq) (Response[[]DictItem], error) {
    if req.DictType == "" {
        req.DictType = "order_type_code"
    }
    if req.IncludeExt == 0 {
        req.IncludeExt = 1
    }
    return agentGET[[]DictItem](getDictItemsEndpoint, req, nil)
}
```

### service.go（追加）

```go
ActionYundaGetDictItems = "yunda.problem.getDictItems"
```

### actions_meta.go（追加）

```go
RegisterActionMeta(ActionMeta{
    Action: ActionYundaGetDictItems, Title: "查询问题件字典", Category: "韵达问题件",
    Description: "查询韵达问题件字典项（默认 dictType=order_type_code）。",
    Encrypted: true, Billable: true, Sort: 100,
})
```

### handler_yunda.go（追加）

```go
// init 中
registerYundaHandler(ActionYundaGetDictItems, validateGetDictItems, yunda.GetDictItems)

// 校验
func validateGetDictItems(req yunda.GetDictItemsReq) error {
    return requireOrgUser(req.OrgCode, req.UserId)
}
```

---

# 示例：multipart 上传 uploadImage

## 用户输入

```
接口: POST https://customer-problem.yundasys.com/upload-img/file/upload

请求体（开放平台 data JSON，非 multipart）:
{
  "filename": "photo.jpg",
  "contentType": "image/jpeg",
  "data": "<base64>"
}
```

## Step 0：action 自动推导

| 项 | 推导过程 | 结果 |
|----|----------|------|
| path | `/upload-img/file/upload` | last=`upload`, prev=`file` |
| method | **路径覆盖表**：`/upload-img/file/upload` → `uploadImage` | `uploadImage` |
| action | | `yunda.problem.uploadImage` |
| Action 常量 | | `ActionYundaUploadImage` |

## 与标准 JSON 接口的差异

1. **types.go**：`UploadImageReq`（须嵌入 `funcapi.AgentRequestBase`）、`UploadImageResponse`
2. **problem.go**：先 `PrepareAgentRequest` 提取 header，再 `DoAgentMultipartRequest`（不是 agentPOST）
3. **handler**：不用 `registerYundaHandler`，单独 `handleYundaUploadImage` 做 base64/大小校验
4. **init**：

```go
ApplyActionTypeSchemas(ActionYundaUploadImage, yunda.UploadImageReq{}, yunda.Response[yunda.UploadImageResponse]{})
registerBusiness(ActionYundaUploadImage, handleYundaUploadImage)
```

### problem.go 要点（multipart + header）

```go
func UploadImage(req UploadImageReq) (Response[UploadImageResponse], error) {
    cleanBody, reqHeaders := funcapi.PrepareAgentRequest(req)
    req = cleanBody.(UploadImageReq)
    // ... 默认值处理 ...
    headers := funcapi.MergeAgentHeaders(defaultProblemHeaders, reqHeaders)
    err := funcapi.DoAgentMultipartRequest(apiBaseURL+uploadImageEndpoint, "POST", parts, &resp, headers)
    return resp, err
}
```

---

# 示例：formData/upload → releaseIssue

## Step 0：路径覆盖

```
POST .../yrh-xt/app/formData/upload
```

| 项 | 结果 |
|----|------|
| 字面推导 | last=`upload` → 可能得 `uploadFormData` |
| **覆盖表** | `/formData/upload` → `releaseIssue` |
| action | `yunda.problem.releaseIssue` |
| Action 常量 | `ActionYundaReleaseIssue` |
| Go 函数 | `ReleaseIssue` |

---

# 示例：分页列表 getProblemManageList

## Step 0：action 自动推导

```
POST .../yrh-xt/app/problemManagementReport/list
```

| 项 | 推导过程 | 结果 |
|----|----------|------|
| 去噪后 | `problemManagementReport`, `list` | |
| method | 规则 B：`get` + shorten(`problemManagementReport`) + `List` | `getProblemManageList` |
| action | | `yunda.problem.getProblemManageList` |

## 代码要点

- 响应 `data` 为分页对象 → `ProblemPage[T]` + `ProblemDetail`
- 函数签名：`Response[ProblemPage[[]ProblemDetail]]`
- 默认值：`PageNum=1`, `PageSize=20`
- POST → `agentPOST`

```go
// ProblemManageReq 须嵌入 funcapi.AgentRequestBase（同 GetDictItemsReq）
type ProblemManageReq struct {
    funcapi.AgentRequestBase
    // ... 业务字段 ...
}

func GetProblemManageList(req ProblemManageReq) (Response[ProblemPage[[]ProblemDetail]], error) {
    if req.PageNum <= 0 {
        req.PageNum = 1
    }
    if req.PageSize <= 0 {
        req.PageSize = 20
    }
    return agentPOST[ProblemPage[[]ProblemDetail]](getProblemManageList, req, nil)
}
```

---

# 示例：TMM 拦截录入 addBusiInterceptBenchmarkData

## Step 0：action 自动推导

```
POST .../tmm/busi_intercept_benchmark/addData
```

| 项 | 结果 |
|----|------|
| host | `tmmweb.yundasys.com:4406` → provider=`tmmweb`, domain=`tmm` |
| **路径覆盖表** | `/busi_intercept_benchmark/addData` → `addBusiInterceptBenchmarkData` |
| action | `tmmweb.tmm.addBusiInterceptBenchmarkData` |

## 代码要点

- `types.go` 第一个字段嵌入 `funcapi.AgentRequestBase`
- `agent.go` 使用 `defaultTmmHeaders` + `PrepareAgentRequest` + `MergeAgentHeaders`
- handler 通过 `registerTmmwebHandler` 注册

```go
// AddBusiInterceptBenchmarkDataReq 拦截录入请求。
type AddBusiInterceptBenchmarkDataReq struct {
    funcapi.AgentRequestBase
    ShipId        string `json:"ship_id"`
    InteType      string `json:"inte_type"`
    StartSiteCode string `json:"start_site_code"`
    // ...
}
```
