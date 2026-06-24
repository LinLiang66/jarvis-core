package openplatform

func init() {
	registerBuiltinActionMetas()
}

func registerBuiltinActionMetas() {
	RegisterActionMetaWithTypes(ActionMeta{
		Action: ActionGetPublicKey, Title: "获取 Token 与公钥", Category: "握手",
		Description: "获取网关会话 token 与服务端 RSA 公钥",
		Encrypted:   false, Billable: false, Sort: 10,
	}, docPublicKeyReq{Timestamp: "", Data: "{}"}, docPublicKeyResp{})

	RegisterActionMetaWithTypes(ActionMeta{
		Action: ActionCreateSecretKey, Title: "3DES 密钥交换", Category: "握手",
		Description: "使用应用 RSA 私钥加密 clientPart，交换 serverPart，合成 3DES 会话密钥",
		Encrypted:   false, Billable: false, Sort: 20,
	}, docSecretKeyReq{}, docSecretKeyResp{})

	RegisterActionMetaWithTypes(ActionMeta{
		Action: ActionEcho, Title: "Echo 演示", Category: "演示",
		Description: "加密回显请求 JSON，用于联调与示例",
		Encrypted:   true, Billable: true, Sort: 30,
	}, docEchoReq{}, docEchoResp{Action: ActionEcho, Message: "pong"})
}
