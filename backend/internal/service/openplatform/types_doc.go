package openplatform

// 握手与演示接口文档实体（用于自动生成 JSON 示例与字段表）。

type docPublicKeyReq struct {
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
}

type docPublicKeyResp struct {
	PublicKey string `json:"publicKey"`
	Token     string `json:"token"`
}

type docSecretKeyReq struct {
	ReqTime string `json:"req_time"`
	Token   string `json:"token"`
	Data    string `json:"data"`
}

type docSecretKeyResp struct {
	ServerPart string `json:"serverPart"`
}

type docEchoReq struct {
	Hello string `json:"hello"`
}

type docEchoResp struct {
	Action  string         `json:"action"`
	Echo    map[string]any `json:"echo"`
	Message string         `json:"message"`
}

type docGatewayWrapResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Success bool   `json:"success"`
}
