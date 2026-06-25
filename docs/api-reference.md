# API 参考

> 管理端接口前缀：`/api/v1`  
> 除登录、健康检查、开放平台网关外，均需请求头：`Authorization: Bearer <token>`

统一响应格式（业务接口）：

```json
{
  "code": 200,
  "message": "success",
  "data": { }
}
```

## 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 服务存活探测 |

## 认证 `/api/v1/auth`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/login` | 登录，返回 token |
| GET | `/userinfo` | 当前用户信息（需登录） |
| POST | `/logout` | 退出（需登录） |

## 系统管理 `/api/v1/system`

### 用户 `/system/user`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/list` | 分页列表 |
| GET | `/:id` | 详情 |
| POST | `` | 新增 |
| PUT | `/:id` | 更新 |
| PUT | `/:id/password` | 重置密码 |
| PUT | `/:id/status` | 更新状态 |
| POST | `/delete` | 批量删除 |

### 角色 `/system/role`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/options` | 下拉选项 |
| GET | `/list` | 分页列表 |
| GET | `/:id` | 详情 |
| POST | `` | 新增 |
| PUT | `/:id` | 更新 |
| GET | `/:id/menus` | 获取角色菜单 |
| PUT | `/:id/menus` | 分配菜单 |
| POST | `/delete` | 批量删除 |

### 菜单 `/system/menu`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/routes` | 当前用户可访问路由（前端动态路由） |
| GET | `/tree` | 菜单树（管理页） |
| POST | `` | 新增 |
| PUT | `/:id` | 更新 |
| POST | `/delete` | 删除 |

### 字典 `/system/dict`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/type/list` | 字典类型列表 |
| POST | `/type` | 新增类型 |
| PUT | `/type/:id` | 更新类型 |
| POST | `/type/delete` | 删除类型 |
| GET | `/data/by-code/:code` | 按编码取字典项 |
| GET | `/data/list` | 字典项列表 |
| POST | `/data` | 新增字典项 |
| PUT | `/data/:id` | 更新字典项 |
| PUT | `/data/:id/status` | 更新状态 |
| POST | `/data/delete` | 删除字典项 |

### 存储配置 `/storage`

> 需超级管理员

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/list` | 列表（`type`：1 本地，2 对象存储） |
| GET | `/:id` | 详情 |
| POST | `` | 新增 |
| PUT | `/:id` | 更新 |
| PUT | `/:id/status` | 启用/禁用 |
| PUT | `/:id/default` | 设为默认存储 |
| POST | `/delete` | 批量删除 |

对象存储字段含 `accessKey`、`secretKey`、`endpoint`、`bucketName`、`baseUrl`（内网访问域名，可选）、`domain`（自定义域名，可选）。

### 文件管理 `/file`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/list` | 分页列表（`parentPath`、`storageId`、`originalName`） |
| GET | `/statistics` | 文件/目录数量与总大小 |
| POST | `/upload` | 上传（`multipart/form-data`：`file`、可选 `parentPath`、`storageId`） |
| POST | `/dir` | 创建文件夹（JSON：`parentPath`、`originalName`） |
| POST | `/delete` | 批量删除（超管；删文件夹会递归删子项及 OSS/本地对象） |

上传图片时若开启 `IMAGE_COMPRESS_*`，服务端自动压缩后再写入存储。

## 静态资源

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/static/{storageCode}/*` | 本地存储文件（无需 JWT） |

## 开放平台管理 `/api/v1/open-app`

> 需管理端登录

### 应用

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/list` | 应用列表 |
| GET | `/:id` | 应用详情 |
| POST | `` | 创建应用（生成 AppID / 密钥） |
| PUT | `/:id` | 更新 |
| POST | `/delete` | 删除 |
| POST | `/:id/regenerate-keys` | 重新生成密钥 |

### 接口 Action

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/action/list` | Action 列表 |
| GET | `/action/by-action` | 按 action 名查询 |
| POST | `/action/sync` | 从代码注册表同步 |
| POST | `/action` | 新增 |
| GET | `/action/:id` | 详情 |
| PUT | `/action/:id` | 更新 |
| POST | `/action/delete` | 删除 |

### 文档与统计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/doc` | 接口文档列表 |
| GET | `/doc/:action` | 单个 Action 文档 |
| GET | `/stat/daily` | 日统计 |
| GET | `/stat/logs` | 调用日志 |

## 开放平台网关

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/open/gateway` | 对外统一网关（`application/x-www-form-urlencoded`） |

网关协议、握手与加解密见 [开放平台](openplatform.md)。
