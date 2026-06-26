package schedulerclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type HandlerFunc func(ctx context.Context, params string) (result string, err error)

type Config struct {
	ServerURL          string
	WorkerToken        string
	InstanceID         string
	WorkerID           string
	Enable             bool
	PollTimeout        time.Duration
	PollEmptyBackoff   time.Duration
	PollIdleSec        time.Duration
}

func ConfigFromEnv() Config {
	pollTimeoutSec := envInt("SCHEDULER_POLL_TIMEOUT_SEC", 30)
	emptyBackoffMs := envInt("SCHEDULER_POLL_EMPTY_BACKOFF_MS", 200)
	pollIdleSec := envInt("SCHEDULER_POLL_IDLE_SEC", 30)
	c := Config{
		ServerURL:        os.Getenv("SCHEDULER_SERVER_URL"),
		WorkerToken:      os.Getenv("SCHEDULER_WORKER_TOKEN"),
		InstanceID:       os.Getenv("SCHEDULER_INSTANCE_ID"),
		WorkerID:         os.Getenv("SCHEDULER_WORKER_ID"),
		Enable:           envBool("SCHEDULER_ENABLE", false),
		PollTimeout:      time.Duration(pollTimeoutSec) * time.Second,
		PollEmptyBackoff: time.Duration(emptyBackoffMs) * time.Millisecond,
		PollIdleSec:      time.Duration(pollIdleSec) * time.Second,
	}
	c.normalizeIDs()
	return c
}

// normalizeIDs fills worker/instance identity when env leaves them empty.
func (c *Config) normalizeIDs() {
	if c.InstanceID == "" {
		c.InstanceID = uuid.NewString()
	}
	if c.WorkerID == "" {
		c.WorkerID = c.InstanceID
	}
}

func (c Config) Enabled() bool {
	return c.Enable && c.ServerURL != "" && c.WorkerToken != ""
}

type Client struct {
	cfg       Config
	handlers  map[string]HandlerFunc
	http      *http.Client
	startOnce sync.Once
}

func New(cfg Config) *Client {
	httpTimeout := cfg.PollTimeout + 15*time.Second
	if httpTimeout < 45*time.Second {
		httpTimeout = 45 * time.Second
	}
	return &Client{
		cfg:      cfg,
		handlers: make(map[string]HandlerFunc),
		http: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

func (c *Client) RegisterHandler(name string, fn HandlerFunc) {
	c.handlers[name] = fn
}

func (c *Client) Start(ctx context.Context) {
	if !c.cfg.Enabled() {
		return
	}
	c.startOnce.Do(func() {
		go c.run(ctx)
	})
}

func (c *Client) run(ctx context.Context) {
	names := c.handlerNames()
	log.Printf("[schedulerclient] starting worker=%s instance=%s handlers=%v", c.cfg.WorkerID, c.cfg.InstanceID, names)
	shouldPoll, err := c.apiRegister(ctx, names)
	if err != nil {
		log.Printf("[schedulerclient] register failed: %v", err)
	}

	hbTicker := time.NewTicker(20 * time.Second)
	defer hbTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-hbTicker.C:
				if err := c.apiHeartbeat(ctx); err != nil {
					log.Printf("[schedulerclient] heartbeat failed: %v", err)
				}
			}
		}
	}()

	idleWait := c.cfg.PollIdleSec
	if idleWait <= 0 {
		idleWait = 30 * time.Second
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !shouldPoll {
			select {
			case <-ctx.Done():
				return
			case <-time.After(idleWait):
			}
			ready, err := c.apiReady(ctx, names)
			if err != nil {
				log.Printf("[schedulerclient] ready check failed: %v", err)
				continue
			}
			shouldPoll = ready
			if !shouldPoll {
				continue
			}
		}

		task, pollEnabled, err := c.apiPoll(ctx, names)
		if err != nil {
			log.Printf("[schedulerclient] poll failed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		shouldPoll = pollEnabled
		if task == nil {
			if !shouldPoll {
				continue
			}
			if backoff := c.cfg.PollEmptyBackoff; backoff > 0 {
				time.Sleep(backoff)
			}
			continue
		}
		c.executeTask(ctx, task)
	}
}

func (c *Client) handlerNames() []string {
	names := make([]string, 0, len(c.handlers))
	for k := range c.handlers {
		names = append(names, k)
	}
	return names
}

type pollTask struct {
	InstanceID int64  `json:"instance_id"`
	JobID      int64  `json:"job_id"`
	Handler    string `json:"handler"`
	Params     string `json:"params"`
	TimeoutSec int    `json:"timeout_sec"`
}

type pollRespData struct {
	pollTask
	ShouldPoll bool `json:"should_poll"`
}

type registerRespData struct {
	ShouldPoll bool `json:"should_poll"`
}

type readyRespData struct {
	ShouldPoll bool `json:"should_poll"`
}

type apiResp struct {
	Code    int             `json:"code"`
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *Client) apiRegister(ctx context.Context, handlers []string) (bool, error) {
	hostname, _ := os.Hostname()
	body := map[string]any{
		"worker_id":   c.cfg.WorkerID,
		"instance_id": c.cfg.InstanceID,
		"hostname":    hostname,
		"handlers":    handlers,
	}
	var data registerRespData
	if err := c.post(ctx, "/worker/v1/register", body, &data); err != nil {
		return false, err
	}
	return data.ShouldPoll, nil
}

func (c *Client) apiHeartbeat(ctx context.Context) error {
	return c.post(ctx, "/worker/v1/heartbeat", map[string]string{"worker_id": c.cfg.WorkerID}, nil)
}

func (c *Client) apiReady(ctx context.Context, handlers []string) (bool, error) {
	var data readyRespData
	if err := c.post(ctx, "/worker/v1/ready", map[string]any{"handlers": handlers}, &data); err != nil {
		return false, err
	}
	return data.ShouldPoll, nil
}

func (c *Client) apiPoll(ctx context.Context, handlers []string) (*pollTask, bool, error) {
	body := map[string]any{
		"worker_id": c.cfg.WorkerID,
		"handlers":  handlers,
	}
	var data pollRespData
	err := c.post(ctx, "/worker/v1/poll", body, &data)
	if err != nil {
		return nil, false, err
	}
	if data.InstanceID == 0 {
		return nil, data.ShouldPoll, nil
	}
	return &data.pollTask, data.ShouldPoll, nil
}

func (c *Client) executeTask(ctx context.Context, task *pollTask) {
	if err := c.post(ctx, "/worker/v1/report/start", map[string]any{
		"worker_id":   c.cfg.WorkerID,
		"instance_id": task.InstanceID,
	}, nil); err != nil {
		log.Printf("[schedulerclient] report start failed: %v", err)
		return
	}
	fn, ok := c.handlers[task.Handler]
	if !ok {
		_ = c.post(ctx, "/worker/v1/report/fail", map[string]any{
			"worker_id":   c.cfg.WorkerID,
			"instance_id": task.InstanceID,
			"error":       "handler not registered: " + task.Handler,
		}, nil)
		return
	}
	timeout := time.Duration(task.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := fn(execCtx, task.Params)
	if err != nil {
		_ = c.post(ctx, "/worker/v1/report/fail", map[string]any{
			"worker_id":   c.cfg.WorkerID,
			"instance_id": task.InstanceID,
			"error":       err.Error(),
		}, nil)
		return
	}
	_ = c.post(ctx, "/worker/v1/report/finish", map[string]any{
		"worker_id":   c.cfg.WorkerID,
		"instance_id": task.InstanceID,
		"result":      result,
	}, nil)
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.ServerURL+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Scheduler-Token", c.cfg.WorkerToken)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeAPIResp(resp.Body, out)
}

func decodeAPIResp(body io.Reader, out any) error {
	respBody, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	var wrap apiResp
	if err := json.Unmarshal(respBody, &wrap); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if wrap.Code != 200 || !wrap.Success {
		return fmt.Errorf("%s", wrap.Message)
	}
	if out != nil && len(wrap.Data) > 0 && string(wrap.Data) != "null" {
		if err := json.Unmarshal(wrap.Data, out); err != nil {
			return err
		}
	}
	return nil
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
