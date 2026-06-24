// Demo: handshake + encrypted Echo (open.demo.echo).
//
// Create an open-platform app in admin UI first, then:
//
//	go run . \
//	  -gateway=http://127.0.0.1:8000/api/v1/open/gateway \
//	  -appid=app_xxx -sign=signSecret -secret=appSecretBase64
//
// Env: OPEN_GATEWAY_URL, OPEN_APP_ID, OPEN_SIGN_SECRET, OPEN_APP_SECRET
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/jarvis/openplatform-go-demo/openplatform"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	gateway := flag.String("gateway", envOr("OPEN_GATEWAY_URL", "http://127.0.0.1:8000/api/v1/open/gateway"), "gateway URL")
	appID := flag.String("appid", envOr("OPEN_APP_ID", ""), "AppID")
	signSecret := flag.String("sign", envOr("OPEN_SIGN_SECRET", ""), "SignSecret")
	appSecret := flag.String("secret", envOr("OPEN_APP_SECRET", ""), "AppSecret (RSA private key DER Base64)")
	flag.Parse()

	if *appID == "" || *signSecret == "" || *appSecret == "" {
		fmt.Println("usage: go run . -appid=xxx -sign=xxx -secret=xxx")
		flag.PrintDefaults()
		os.Exit(1)
	}

	client, err := openplatform.NewClient(*gateway, *appID, *signSecret, *appSecret)
	if err != nil {
		panic(err)
	}

	token, err := client.GetPublicKeyAndToken()
	if err != nil {
		panic(err)
	}
	fmt.Println("token:", token)

	if err := client.Init3DesKey(); err != nil {
		panic(err)
	}
	fmt.Println("3des session ready")

	echoPayload, _ := json.Marshal(map[string]string{
		"hello": "jarvis demo",
	})
	echoResp, err := client.CallEncrypted(openplatform.ActionEcho, string(echoPayload))
	if err != nil {
		panic(err)
	}
	fmt.Println("echo response:", echoResp)
}
