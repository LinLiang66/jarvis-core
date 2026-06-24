package openplatform

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jarvis/backend/internal/pkg/logx"
	opsvc "jarvis/backend/internal/service/openplatform"
)

type GatewayHandler struct {
	svc *opsvc.Service
}

func NewGatewayHandler(svc *opsvc.Service) *GatewayHandler {
	return &GatewayHandler{svc: svc}
}

func (h *GatewayHandler) Register(rg *gin.RouterGroup) {
	rg.Use(accessLogMiddleware("gateway"))
	rg.POST("/gateway", h.Gateway)
}

// Gateway 开放平台统一入口（application/x-www-form-urlencoded）。
func (h *GatewayHandler) Gateway(c *gin.Context) {
	params := make(map[string]string)
	if err := c.Request.ParseForm(); err != nil {
		logx.Infof("[openplatform][gateway] parse form failed ip=%s err=%v", c.ClientIP(), err)
		c.JSON(http.StatusOK, opsvc.DecoySuccessResponse())
		return
	}
	for k, vals := range c.Request.PostForm {
		if len(vals) > 0 {
			params[k] = vals[0]
		}
	}
	req := h.svc.ParseForm(params)
	logx.Infof("[openplatform][gateway] request action=%s appid=%s token_len=%d data_len=%d",
		req.Action, req.AppID, len(req.Token), len(req.Data))
	resp, _ := h.svc.Handle(c.Request.Context(), req, c.ClientIP())
	logx.Infof("[openplatform][gateway] response action=%s code=%d message=%s has_data=%v",
		req.Action, resp.Code, resp.Message, resp.Data != nil)
	c.JSON(http.StatusOK, resp)
}
