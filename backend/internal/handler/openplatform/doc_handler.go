package openplatform

import (
	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/pkg/response"
	opsvc "jarvis-core/backend/internal/service/openplatform"
)

type DocHandler struct {
	svc *opsvc.Service
}

func NewDocHandler(svc *opsvc.Service) *DocHandler {
	return &DocHandler{svc: svc}
}

func (h *DocHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/doc")
	g.Use(accessLogMiddleware("open-action-doc"))
	g.GET("", h.List)
	g.GET("/:action", h.Detail)
}

// List 接口文档目录（需登录）。
func (h *DocHandler) List(c *gin.Context) {
	list, err := h.svc.ListPublicDoc(c.Request.Context())
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

// Detail 按 action 查询单个接口文档。
func (h *DocHandler) Detail(c *gin.Context) {
	action := c.Param("action")
	if action == "" {
		response.Fail(c, 400, "action required")
		return
	}
	item, err := h.svc.GetPublicDocByAction(c.Request.Context(), action)
	if err != nil {
		response.Fail(c, 404, "接口不存在或已禁用")
		return
	}
	response.OK(c, item)
}
