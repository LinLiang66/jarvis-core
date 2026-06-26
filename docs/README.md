# 项目文档

jarvis-core 是一套前后端分离的管理后台基础框架，内置 **系统管理**（用户 / 角色 / 菜单 / 字典 / 存储配置 / 文件管理）与 **开放平台**（网关、应用、接口文档、调用统计），可作为业务系统的起点快速扩展。

## 文档目录

| 文档 | 说明 |
|------|------|
| [快速开始](getting-started.md) | 环境准备、本地启动、默认账号 |
| [架构说明](architecture.md) | 目录结构、模块划分、认证与权限 |
| [API 参考](api-reference.md) | 管理端 REST 接口一览 |
| [开放平台](openplatform.md) | 网关协议、握手流程、接入与扩展 |
| [部署指南](deployment.md) | Docker、生产环境、前后端分离部署 |
| [任务调度](scheduler.md) | scheduler-server、Worker、BFF 代理与配置 |
| [开发指南](development.md) | 二次开发、菜单扩展、新增 Action |

## 子项目说明

| 路径 | 文档 |
|------|------|
| 后端 | [backend/README.md](../backend/README.md) |
| 调度服务 | [scheduler/.env.example](../scheduler/.env.example) · [任务调度](scheduler.md) |
| 前端 | [frontend/web/README.md](../frontend/web/README.md) |
| Docker | [docker/README.md](../docker/README.md) |
| SDK 示例 | [examples/README.md](../examples/README.md) |

## 仓库与上游

| 远程 | 地址 | 用途 |
|------|------|------|
| origin | https://gitcode.com/LinLiang/jarvis-core.git | 开源基础框架（master） |
| upstream | https://github.com/lin-97/gi-element-plus-admin.git | 上游 UI 框架（main） |

同步上游 UI 框架：

```bash
git fetch upstream
git merge upstream/main
```
