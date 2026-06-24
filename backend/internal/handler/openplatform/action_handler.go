package openplatform

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/pkg/parseid"
	"jarvis-core/backend/internal/pkg/response"
	opsvc "jarvis-core/backend/internal/service/openplatform"
	"jarvis-core/backend/internal/store"
)

type ActionHandler struct {
	svc *opsvc.Service
}

func NewActionHandler(svc *opsvc.Service) *ActionHandler {
	return &ActionHandler{svc: svc}
}

func (h *ActionHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/action")
	g.Use(accessLogMiddleware("open-action"))
	g.GET("/list", h.List)
	g.GET("/by-action", h.GetByAction)
	g.POST("/sync", h.Sync)
	g.POST("", h.Create)
	g.GET("/:id", h.Detail)
	g.PUT("/:id", h.Update)
	g.POST("/delete", h.Delete)
}

func (h *ActionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.svc.ListActions(c.Request.Context(), store.PageQuery{Page: page, Size: size}, store.OpenAPIActionFilter{
		Action:   c.Query("action"),
		Title:    c.Query("title"),
		Category: c.Query("category"),
		Status:   c.Query("status"),
	})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, list, int(total), page, size)
}

func (h *ActionHandler) Detail(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	idNum, _ := parseid.Path(id)
	item, err := h.svc.GetActionByID(c.Request.Context(), idNum)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	response.OK(c, item)
}

func (h *ActionHandler) GetByAction(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		response.Fail(c, 400, "action required")
		return
	}
	item, err := h.svc.GetActionByAction(c.Request.Context(), action)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	response.OK(c, item)
}

func (h *ActionHandler) Sync(c *gin.Context) {
	n, err := h.svc.SyncActionRegistry(c.Request.Context())
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"synced": n})
}

type createActionReq struct {
	Action         string `json:"action" binding:"required"`
	Title          string `json:"title" binding:"required"`
	Category       string `json:"category"`
	Description    string `json:"description"`
	Encrypted      bool   `json:"encrypted"`
	Billable       bool   `json:"billable"`
	Status         string `json:"status"`
	RequestSchema  string `json:"request_schema"`
	ResponseSchema string `json:"response_schema"`
	Sort           int    `json:"sort"`
}

func (h *ActionHandler) Create(c *gin.Context) {
	var req createActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	status := req.Status
	if status == "" {
		status = "0"
	}
	row := &model.OpenAPIAction{
		Action:         req.Action,
		Title:          req.Title,
		Category:       req.Category,
		Description:    req.Description,
		Encrypted:      req.Encrypted,
		Billable:       req.Billable,
		Status:         status,
		RequestSchema:  req.RequestSchema,
		ResponseSchema: req.ResponseSchema,
		Sort:           req.Sort,
	}
	if err := h.svc.CreateAction(c.Request.Context(), row); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, row)
}

func (h *ActionHandler) Update(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	idNum, _ := parseid.Path(id)
	item, err := h.svc.GetActionByID(c.Request.Context(), idNum)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	var patch struct {
		Title          *string `json:"title"`
		Category       *string `json:"category"`
		Description    *string `json:"description"`
		Encrypted      *bool   `json:"encrypted"`
		Billable       *bool   `json:"billable"`
		Status         *string `json:"status"`
		RequestSchema  *string `json:"request_schema"`
		ResponseSchema *string `json:"response_schema"`
		Sort           *int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&patch); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if patch.Title != nil {
		item.Title = *patch.Title
	}
	if patch.Category != nil {
		item.Category = *patch.Category
	}
	if patch.Description != nil {
		item.Description = *patch.Description
	}
	if patch.Encrypted != nil {
		item.Encrypted = *patch.Encrypted
	}
	if patch.Billable != nil {
		item.Billable = *patch.Billable
	}
	if patch.Status != nil {
		item.Status = *patch.Status
	}
	if patch.RequestSchema != nil {
		item.RequestSchema = *patch.RequestSchema
	}
	if patch.ResponseSchema != nil {
		item.ResponseSchema = *patch.ResponseSchema
	}
	if patch.Sort != nil {
		item.Sort = *patch.Sort
	}
	if err := h.svc.UpdateAction(c.Request.Context(), item); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *ActionHandler) Delete(c *gin.Context) {
	ids, err := parseid.BindDeleteIDs(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if err := h.svc.DeleteActions(c.Request.Context(), ids); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}
