package scheduler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/pkg/response"
)

type Handler struct {
	cfg    *config.Config
	client *http.Client
}

func New(cfg *config.Config) *Handler {
	return &Handler{
		cfg:    cfg,
		client: &http.Client{},
	}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	if !h.cfg.SchedulerEnabled() {
		return
	}
	rg.Any("/*path", h.Proxy)
}

func (h *Handler) Proxy(c *gin.Context) {
	if h.cfg.SchedulerServerURL == "" {
		response.Fail(c, 503, "调度服务未配置")
		return
	}
	subPath := strings.TrimPrefix(c.Param("path"), "/")
	target := strings.TrimRight(h.cfg.SchedulerServerURL, "/") + "/admin/v1/" + subPath
	if c.Request.URL.RawQuery != "" {
		target += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, target, c.Request.Body)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	for k, vals := range c.Request.Header {
		if strings.EqualFold(k, "Authorization") || strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("X-Scheduler-Token", h.cfg.SchedulerAdminToken)

	resp, err := h.client.Do(req)
	if err != nil {
		response.Fail(c, 502, "调度服务不可用: "+err.Error())
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		if strings.EqualFold(k, "Content-Length") {
			continue
		}
		for _, v := range vals {
			c.Header(k, v)
		}
	}
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}
