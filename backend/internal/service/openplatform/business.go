package openplatform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/pkg/logx"
)

// BusinessHandler 3DES 加密业务 action 处理器。
type BusinessHandler func(ctx context.Context, app *model.OpenApp, plain []byte) ([]byte, error)

var businessHandlers = map[string]BusinessHandler{}

func registerBusiness(action string, h BusinessHandler) {
	EnsureActionMeta(action, "业务", true, true)
	businessHandlers[action] = h
}

func init() {
	registerBusiness(ActionEcho, handleEcho)
}

func (s *Service) handleBusiness(ctx context.Context, app *model.OpenApp, action string, plain []byte) ([]byte, error) {
	h, ok := businessHandlers[action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", action)
	}
	logx.Infof("[openplatform] business >>> app=%s action=%s req=%s",
		app.AppID, action, truncate(string(plain), 1000))
	result, err := h(ctx, app, plain)
	if err != nil {
		logx.Infof("[openplatform] business <<< app=%s action=%s failed err=%v",
			app.AppID, action, err)
		return nil, err
	}
	logx.Infof("[openplatform] business <<< app=%s action=%s resp=%s",
		app.AppID, action, truncate(string(result), 1000))
	return result, nil
}

func handleEcho(_ context.Context, _ *model.OpenApp, plain []byte) ([]byte, error) {
	var in map[string]any
	if err := json.Unmarshal(plain, &in); err != nil {
		return nil, errors.New("invalid json in data")
	}
	return json.Marshal(map[string]any{
		"action":  ActionEcho,
		"echo":    in,
		"message": "pong",
	})
}
