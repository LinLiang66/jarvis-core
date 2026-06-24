# 开放平台 SDK 示例

演示如何通过统一网关调用 jarvis-core 开放平台（握手、3DES 加密、Echo 演示接口）。

> 示例**仅调用内置演示 Action** `open.demo.echo`，不包含任何真实业务系统接口，避免泄露业务 Action 名称或参数结构。

## 目录

| 示例 | 语言 | 路径 |
|------|------|------|
| Go | Go 1.21+ | [openplatform-go-demo/](openplatform-go-demo/) |
| Python | Python 3.10+ | [openplatform-python-demo/](openplatform-python-demo/) |
| Java | Java 11+ | [openplatform-java-demo/](openplatform-java-demo/) |

## 前置条件

1. jarvis-core 后端已启动（默认 `http://127.0.0.1:8000`）
2. 管理端创建开放平台应用，记录 **AppID**、**SignSecret**、**AppSecret**（RSA 私钥 DER Base64）
3. 管理端 **接口 → 同步**，并为应用授权 **`open.demo.echo`**

## 调用流程

1. `open.session.publickey` — 获取 token
2. `microSession.create.secretkey` — 3DES 密钥交换
3. `open.demo.echo` — 加密 Echo 演示（回显请求 JSON）

## Go

```powershell
cd openplatform-go-demo
go run . -appid=app_xxx -sign=signSecret -secret=appSecretBase64
```

环境变量：`OPEN_GATEWAY_URL`、`OPEN_APP_ID`、`OPEN_SIGN_SECRET`、`OPEN_APP_SECRET`

## Python

```powershell
cd openplatform-python-demo
pip install -r requirements.txt
python demo.py --appid app_xxx --sign signSecret --secret appSecretBase64
```

## Java

```powershell
cd openplatform-java-demo
mvn -q exec:java -Dexec.args="http://127.0.0.1:8000/api/v1/open/gateway app_xxx signSecret appSecretBase64"
```

协议细节见 [docs/openplatform.md](../docs/openplatform.md)。
