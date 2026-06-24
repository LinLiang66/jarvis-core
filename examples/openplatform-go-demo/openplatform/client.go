package openplatform

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jarvis/openplatform-go-demo/crypto"
)

const (
	ActionGetPublicKey    = "open.session.publickey"
	ActionCreateSecretKey = "microSession.create.secretkey"
	ActionEcho            = "open.demo.echo"

	appVer  = "1.0.0"
	version = "V1.0"
)

type gatewayResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Success bool            `json:"success"`
}

// Client is the open-platform SDK: handshake + 3DES encrypted calls.
type Client struct {
	gatewayURL string
	appID      string
	signSecret string
	privatePEM string
	httpClient *http.Client

	token      string
	tdesCipher *crypto.TDESCipher
}

func NewClient(gatewayURL, appID, signSecret, appSecretBase64 string) (*Client, error) {
	privPEM, err := crypto.LoadPrivateKeyFromDerBase64(appSecretBase64)
	if err != nil {
		return nil, err
	}
	return &Client{
		gatewayURL: gatewayURL,
		appID:      appID,
		signSecret: signSecret,
		privatePEM: privPEM,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetPublicKeyAndToken step 1: obtain token.
func (c *Client) GetPublicKeyAndToken() (string, error) {
	params := c.baseParams(ActionGetPublicKey)
	params["timestamp"] = fmt.Sprintf("%d", time.Now().UnixMilli())
	params["data"] = "{}"

	var body struct {
		Token string `json:"token"`
	}
	if err := c.callGateway(params, &body); err != nil {
		return "", err
	}
	c.token = body.Token
	return c.token, nil
}

// Init3DesKey step 2: 3DES key exchange.
func (c *Client) Init3DesKey() error {
	if c.token == "" {
		return fmt.Errorf("call GetPublicKeyAndToken first")
	}
	randomNum, err := crypto.RandomDigits(12)
	if err != nil {
		return err
	}
	encrypted, err := crypto.EncryptByPrivateKey(c.privatePEM, randomNum)
	if err != nil {
		return err
	}
	params := c.baseParams(ActionCreateSecretKey)
	params["req_time"] = fmt.Sprintf("%d", time.Now().UnixMilli())
	params["token"] = c.token
	params["data"] = url.QueryEscape(encrypted)

	var body struct {
		ServerPart string `json:"serverPart"`
	}
	if err := c.callGateway(params, &body); err != nil {
		return err
	}
	serverPart, err := crypto.DecryptByPrivateKey(c.privatePEM, body.ServerPart)
	if err != nil {
		return err
	}
	finalKey := randomNum + serverPart
	c.tdesCipher, err = crypto.NewTDESCipher([]byte(finalKey))
	return err
}

// CallEncrypted step 3: encrypted business call, returns decrypted JSON string.
func (c *Client) CallEncrypted(action, jsonPlain string) (string, error) {
	if c.tdesCipher == nil {
		return "", fmt.Errorf("call Init3DesKey first")
	}
	cipherData, err := c.tdesCipher.Encrypt([]byte(jsonPlain))
	if err != nil {
		return "", err
	}
	params := c.baseParams(action)
	params["req_time"] = fmt.Sprintf("%d", time.Now().UnixMilli())
	params["token"] = c.token
	params["data"] = cipherData

	var raw json.RawMessage
	if err := c.callGatewayRaw(params, &raw); err != nil {
		return "", err
	}
	var cipherBody string
	if err := json.Unmarshal(raw, &cipherBody); err != nil {
		cipherBody = string(raw)
	}
	plain, err := c.tdesCipher.Decrypt(cipherBody)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (c *Client) Token() string { return c.token }

func (c *Client) baseParams(action string) map[string]string {
	return map[string]string{
		"action":      action,
		"appid":       c.appID,
		"appver":      appVer,
		"version":     version,
		"sign_method": crypto.SignMethodA2MD5,
	}
}

func (c *Client) callGateway(params map[string]string, out any) error {
	return c.callGatewayRaw(params, out)
}

func (c *Client) callGatewayRaw(params map[string]string, out any) error {
	signed := make(map[string]string, len(params)+1)
	for k, v := range params {
		signed[k] = v
	}
	signed["sign"] = crypto.BuildA2MD5Sign(signed, c.signSecret)

	form := url.Values{}
	for k, v := range signed {
		form.Set(k, v)
	}
	req, err := http.NewRequest(http.MethodPost, c.gatewayURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var resp gatewayResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return err
	}
	if resp.Code == CodeTokenInvalid {
		msg := resp.Message
		if msg == "" {
			msg = "token invalid or expired"
		}
		return &OpenPlatformError{Code: CodeTokenInvalid, Message: msg}
	}
	if resp.Code == CodeQuotaExceeded {
		msg := resp.Message
		if msg == "" {
			msg = "quota exceeded"
		}
		return &OpenPlatformError{Code: CodeQuotaExceeded, Message: msg}
	}
	if resp.Code != CodeSuccess {
		return &OpenPlatformError{Code: resp.Code, Message: fmt.Sprintf("gateway error: code=%d message=%s", resp.Code, resp.Message)}
	}
	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return &OpenPlatformError{Code: CodeSuccess, Message: "empty response data"}
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(resp.Data, out)
}
