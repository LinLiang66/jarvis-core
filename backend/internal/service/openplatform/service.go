package openplatform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/pkg/crypto"
	"jarvis-core/backend/internal/pkg/logx"
	"jarvis-core/backend/internal/store"
)

const (
	ActionGetPublicKey                        = "open.session.publickey"
	ActionCreateSecretKey                     = "microSession.create.secretkey"
	ActionEcho            = "open.demo.echo"
	DefaultVersion        = "V1.0"
)

// GatewayRequest å¼æ¾å¹³å°ç½å³è¯·æ±åæ°ï¼form-urlencodedï¼ã?
type GatewayRequest struct {
	Action     string
	AppID      string
	AppVer     string
	Version    string
	SignMethod string
	Sign       string
	Token      string
	Data       string
	Timestamp  string
	ReqTime    string
}

// GatewayResponse å¼æ¾å¹³å°ç½å³ååºã?
type GatewayResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Success bool   `json:"success"`
}

type Service struct {
	apps       *store.OpenAppRepository
	stats      *store.OpenAPIStatRepository
	actions    *store.OpenAPIActionRepository
	statStore  *StatStore
	quotaStore *QuotaStore
	session    *SessionStore
}

func NewService(apps *store.OpenAppRepository, stats *store.OpenAPIStatRepository, actions *store.OpenAPIActionRepository, statStore *StatStore, quotaStore *QuotaStore, session *SessionStore) *Service {
	return &Service{apps: apps, stats: stats, actions: actions, statStore: statStore, quotaStore: quotaStore, session: session}
}

func (s *Service) ParseForm(params map[string]string) GatewayRequest {
	return GatewayRequest{
		Action:     params["action"],
		AppID:      params["appid"],
		AppVer:     params["appver"],
		Version:    params["version"],
		SignMethod: params["sign_method"],
		Sign:       params["sign"],
		Token:      params["token"],
		Data:       params["data"],
		Timestamp:  params["timestamp"],
		ReqTime:    params["req_time"],
	}
}

func (s *Service) Handle(ctx context.Context, req GatewayRequest, clientIP string) (*GatewayResponse, error) {
	start := time.Now()
	logx.Infof("[openplatform] >>> app=%s action=%s ip=%s token=%s data_len=%d",
		req.AppID, req.Action, clientIP, maskToken(req.Token), len(req.Data))

	ctx, resp, err := s.dispatch(ctx, req)
	success := err == nil
	s.settleQuota(ctx, req.Action, success)
	msg := "success"
	if err != nil {
		msg = err.Error()
	}
	duration := time.Since(start).Milliseconds()
	if req.AppID != "" && req.Action != "" && shouldRecordStat(req.Action) {
		callLog := &model.OpenAPICallLog{
			AppID:    req.AppID,
			Action:   req.Action,
			Success:  success,
			Duration: duration,
			Message:  truncate(msg, 500),
			ClientIP: clientIP,
		}
		_ = s.stats.RecordCallLog(ctx, callLog)
		_ = s.statStore.Record(ctx, req.AppID, req.Action, success)
	}
	if err != nil {
		errResp := BuildGatewayErrorResponse(err)
		logx.Infof("[openplatform] <<< app=%s action=%s success=false duration=%dms err=%s code=%d decoy=%v",
			req.AppID, req.Action, duration, msg, errResp.Code, isDecoyResponse(errResp))
		return errResp, nil
	}
	logx.Infof("[openplatform] <<< app=%s action=%s success=true duration=%dms code=%d has_data=%v",
		req.AppID, req.Action, duration, resp.Code, resp.Data != nil)
	return resp, nil
}

func (s *Service) dispatch(ctx context.Context, req GatewayRequest) (context.Context, *GatewayResponse, error) {
	switch req.Action {
	case ActionGetPublicKey:
		return s.getPublicKey(ctx, req)
	case ActionCreateSecretKey:
		return s.createSecretKey(ctx, req)
	default:
		return s.handleEncrypted(ctx, req)
	}
}

func (s *Service) getPublicKey(ctx context.Context, req GatewayRequest) (context.Context, *GatewayResponse, error) {
	app, err := s.verifySign(ctx, req)
	if err != nil {
		return ctx, nil, err
	}
	ctx, err = s.reserveQuota(ctx, app.AppID, req.Action)
	if err != nil {
		return ctx, nil, err
	}
	token, err := newToken()
	if err != nil {
		return ctx, nil, err
	}
	info := &SessionInfo{
		AppID:     app.AppID,
		Token:     token,
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := s.session.SaveToken(ctx, info); err != nil {
		return ctx, nil, err
	}
	body := map[string]string{
		"publicKey": crypto.StripPEMHeaders(app.RSAPublicKey),
		"token":     token,
	}
	return ctx, &GatewayResponse{Code: CodeSuccess, Message: "success", Data: body, Success: true}, nil
}

func (s *Service) createSecretKey(ctx context.Context, req GatewayRequest) (context.Context, *GatewayResponse, error) {
	app, err := s.verifySign(ctx, req)
	if err != nil {
		return ctx, nil, err
	}
	if req.Token == "" {
		return ctx, nil, errors.New("token required")
	}
	sess, err := s.session.GetByToken(ctx, req.Token)
	if err != nil || sess.AppID != app.AppID {
		return ctx, nil, ErrTokenInvalid
	}
	ctx, err = s.reserveQuota(ctx, app.AppID, req.Action)
	if err != nil {
		return ctx, nil, err
	}
	encryptedData, err := url.QueryUnescape(req.Data)
	if err != nil {
		encryptedData = req.Data
	}
	clientPart, err := crypto.DecryptByPublicKey(app.RSAPublicKey, encryptedData)
	if err != nil {
		return ctx, nil, fmt.Errorf("decrypt client part: %w", err)
	}
	serverPart, err := crypto.RandomDigits(12)
	if err != nil {
		return ctx, nil, err
	}
	serverCipher, err := crypto.EncryptByPublicKey(app.RSAPublicKey, serverPart)
	if err != nil {
		return ctx, nil, err
	}
	finalKey := clientPart + serverPart
	if err := s.session.SaveTDESKey(ctx, req.Token, finalKey); err != nil {
		return ctx, nil, err
	}
	body := map[string]string{"serverPart": serverCipher}
	return ctx, &GatewayResponse{Code: CodeSuccess, Message: "success", Data: body, Success: true}, nil
}

func (s *Service) handleEncrypted(ctx context.Context, req GatewayRequest) (context.Context, *GatewayResponse, error) {
	app, err := s.verifySign(ctx, req)
	if err != nil {
		return ctx, nil, err
	}
	if req.Token == "" {
		return ctx, nil, errors.New("token required")
	}
	sess, err := s.session.GetByToken(ctx, req.Token)
	if err != nil || sess.AppID != app.AppID {
		return ctx, nil, ErrTokenInvalid
	}
	s.session.TouchSession(ctx, req.Token)
	cipher, err := s.session.GetCryptor(ctx, req.Token)
	if err != nil {
		return ctx, nil, errors.New("3des key not initialized, call microSession.create.secretkey first")
	}
	ctx, err = s.reserveQuota(ctx, app.AppID, req.Action)
	if err != nil {
		return ctx, nil, err
	}
	plain, err := cipher.Decrypt(req.Data)
	if err != nil {
		return ctx, nil, fmt.Errorf("decrypt request: %w", err)
	}
	result, err := s.handleBusiness(ctx, app, req.Action, plain)
	if err != nil {
		return ctx, nil, err
	}
	cipherText, err := cipher.Encrypt(result)
	if err != nil {
		return ctx, nil, err
	}
	return ctx, &GatewayResponse{Code: CodeSuccess, Message: "success", Data: cipherText, Success: true}, nil
}

func (s *Service) verifySign(ctx context.Context, req GatewayRequest) (*model.OpenApp, error) {
	if req.AppID == "" {
		return nil, errors.New("appid required")
	}
	if req.SignMethod != "" && req.SignMethod != crypto.SignMethodA2MD5 {
		return nil, errors.New("unsupported sign_method")
	}
	app, err := s.apps.GetByAppID(ctx, req.AppID)
	if err != nil {
		return nil, errors.New("app not found or disabled")
	}
	params := map[string]string{
		"action":      req.Action,
		"appid":       req.AppID,
		"appver":      req.AppVer,
		"version":     req.Version,
		"sign_method": req.SignMethod,
		"token":       req.Token,
		"data":        req.Data,
		"timestamp":   req.Timestamp,
		"req_time":    req.ReqTime,
	}
	if !crypto.VerifyA2MD5Sign(params, app.SignSecret, req.Sign) {
		return nil, errors.New("invalid sign")
	}
	return app, nil
}

type quotaCtxKey struct{}

// isHandshakeAction æ¡æç±»æ¥å£ï¼ä¸è®¡è´¹ãä¸çº³å¥è°ç¨æ¬¡æ°ç»è®¡ã?
func isHandshakeAction(action string) bool {
	return action == ActionGetPublicKey || action == ActionCreateSecretKey
}

func shouldChargeQuota(action string) bool {
	return !isHandshakeAction(action)
}

func shouldRecordStat(action string) bool {
	return !isHandshakeAction(action)
}

func (s *Service) reserveQuota(ctx context.Context, appID, action string) (context.Context, error) {
	if !shouldChargeQuota(action) {
		return ctx, nil
	}
	if s.quotaStore != nil && s.quotaStore.Enabled() {
		ok, err := s.quotaStore.TryDeduct(ctx, appID)
		if err != nil {
			return ctx, err
		}
		if !ok {
			return ctx, ErrQuotaExceeded
		}
		return context.WithValue(ctx, quotaCtxKey{}, appID), nil
	}
	ok, err := s.apps.TryDeductQuota(ctx, appID)
	if err != nil {
		return ctx, err
	}
	if !ok {
		return ctx, ErrQuotaExceeded
	}
	return context.WithValue(ctx, quotaCtxKey{}, appID), nil
}

func (s *Service) settleQuota(ctx context.Context, action string, success bool) {
	appID, _ := ctx.Value(quotaCtxKey{}).(string)
	if appID == "" || !shouldChargeQuota(action) {
		return
	}
	if s.quotaStore != nil && s.quotaStore.Enabled() {
		if success {
			_ = s.apps.IncrTotalCalls(ctx, appID, 1)
			_ = s.quotaStore.FlushBalanceToDB(ctx, appID)
		} else {
			_ = s.quotaStore.Refund(ctx, appID)
		}
		return
	}
	if success {
		_ = s.apps.ConfirmCall(ctx, appID)
		return
	}
	_ = s.apps.RefundQuota(ctx, appID)
}

// SyncQuotaToRedis ç®¡çç«¯è°æ´éé¢ååæ­¥ Redis ä½é¢ã?
func (s *Service) SyncQuotaToRedis(ctx context.Context, appID string, balance int) error {
	if s.quotaStore == nil {
		return nil
	}
	return s.quotaStore.SyncBalanceFromDB(ctx, appID, balance)
}

// CreateApp åå»ºå¼æ¾å¹³å°åºç¨ï¼è¿åå«ç§é¥çå®æ´ä¿¡æ¯ï¼ä»åå»ºæ¶æ´é²ç§é¥ï¼ã?
func (s *Service) CreateApp(ctx context.Context, name string, totalQuota int, remark string) (*model.OpenApp, string, error) {
	privPEM, pubPEM, err := crypto.GenerateRSAKeyPair()
	if err != nil {
		return nil, "", err
	}
	appID, err := genAppID()
	if err != nil {
		return nil, "", err
	}
	signSecret, err := genSignSecret()
	if err != nil {
		return nil, "", err
	}
	app := &model.OpenApp{
		AppID:         appID,
		AppName:       name,
		SignSecret:    signSecret,
		RSAPublicKey:  pubPEM,
		RSAPrivateKey: privPEM,
		Status:        "0",
		TotalQuota:    totalQuota,
		Remark:        remark,
	}
	if err := s.apps.Create(ctx, app); err != nil {
		return nil, "", err
	}
	_ = s.SyncQuotaToRedis(ctx, app.AppID, app.TotalQuota)
	logx.Infof("[openplatform] app created app_id=%s name=%s quota=%d", app.AppID, app.AppName, app.TotalQuota)
	// appSecret = RSA ç§é¥ base64(DER)ï¼å®¢æ·ç«¯ç¨äº RSA å è§£å¯?
	appSecret := crypto.StripPEMHeaders(privPEM)
	return app, appSecret, nil
}

func (s *Service) RegenerateKeys(ctx context.Context, id int64) (*model.OpenApp, string, error) {
	app, err := s.apps.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	privPEM, pubPEM, err := crypto.GenerateRSAKeyPair()
	if err != nil {
		return nil, "", err
	}
	signSecret, err := genSignSecret()
	if err != nil {
		return nil, "", err
	}
	app.RSAPublicKey = pubPEM
	app.RSAPrivateKey = privPEM
	app.SignSecret = signSecret
	if err := s.apps.Save(ctx, app); err != nil {
		return nil, "", err
	}
	logx.Infof("[openplatform] app keys regenerated app_id=%s id=%d", app.AppID, id)
	return app, crypto.StripPEMHeaders(privPEM), nil
}

func genAppID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "app_" + hex.EncodeToString(b), nil
}

func genSignSecret() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func newToken() (string, error) {
	return strings.ReplaceAll(uuid.NewString(), "-", ""), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func maskToken(token string) string {
	if token == "" {
		return "-"
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
