package schedulerclient

import (
	"context"
	"fmt"
	"log"
	"time"

	"jarvis-core/backend/internal/pkg/logx"
)

// RegisterStatSyncHandler 注册 stat.sync 任务处理器（MVP stub）。
func RegisterStatSyncHandler(c *Client) {
	c.RegisterHandler("stat.sync", func(ctx context.Context, params string) (string, error) {
		logx.Infof("[schedulerclient] stat.sync executed params=%s", params)
		return `{"ok":true,"message":"stat sync stub completed"}`, nil
	})
}

// RegisterDemoHelloHandler 注册 demo.hello 示例任务处理器。
func RegisterDemoHelloHandler(c *Client) {
	c.RegisterHandler("demo.hello", func(ctx context.Context, params string) (string, error) {
		now := time.Now().Format(time.RFC3339)
		logx.Infof("[schedulerclient] demo.hello executed at=%s params=%s", now, params)
		return fmt.Sprintf(`{"ok":true,"message":"hello at %s"}`, now), nil
	})
}

// StartDefault 启动默认 worker 并注册内置 handler。
func StartDefault(ctx context.Context, cfg Config) *Client {
	client := New(cfg)
	RegisterStatSyncHandler(client)
	RegisterDemoHelloHandler(client)
	client.Start(ctx)
	if cfg.Enabled() {
		log.Printf("[schedulerclient] worker enabled url=%s", cfg.ServerURL)
	}
	return client
}
