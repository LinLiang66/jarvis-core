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

type AdminHandler struct {
	apps  *store.OpenAppRepository
	stats *store.OpenAPIStatRepository
	svc   *opsvc.Service
}

func NewAdminHandler(stores *store.Stores, svc *opsvc.Service) *AdminHandler {
	return &AdminHandler{
		apps:  stores.OpenApp,
		stats: stores.OpenAPIStat,
		svc:   svc,
	}
}

func (h *AdminHandler) Register(rg *gin.RouterGroup) {
	rg.Use(accessLogMiddleware("admin"))
	rg.GET("/stat/daily", h.ListDailyStat)
	rg.GET("/stat/logs", h.ListCallLogs)
	rg.GET("/list", h.List)
	rg.GET("/:id", h.Detail)
	rg.POST("", h.Create)
	rg.PUT("/:id", h.Update)
	rg.POST("/delete", h.Delete)
	rg.POST("/:id/regenerate-keys", h.RegenerateKeys)
}

type createAppReq struct {
	AppName    string `json:"app_name" binding:"required"`
	TotalQuota int    `json:"total_quota"`
	Remark     string `json:"remark"`
}

type createAppResp struct {
	model.OpenApp
	AppSecret  string `json:"app_secret"`
	SignSecret string `json:"sign_secret"`
}

func (h *AdminHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.apps.List(c.Request.Context(), store.PageQuery{Page: page, Size: size}, store.OpenAppFilter{
		AppID:   c.Query("app_id"),
		AppName: c.Query("app_name"),
		Status:  c.Query("status"),
	})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, list, int(total), page, size)
}

func (h *AdminHandler) Detail(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	item, err := h.apps.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	response.OK(c, item)
}

func (h *AdminHandler) Create(c *gin.Context) {
	var req createAppReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	app, appSecret, err := h.svc.CreateApp(c.Request.Context(), req.AppName, req.TotalQuota, req.Remark)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, createAppResp{
		OpenApp:    *app,
		AppSecret:  appSecret,
		SignSecret: app.SignSecret,
	})
}

func (h *AdminHandler) Update(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	item, err := h.apps.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	var patch struct {
		AppName    *string `json:"app_name"`
		Status     *string `json:"status"`
		TotalQuota *int    `json:"total_quota"`
		Remark     *string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&patch); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if patch.AppName != nil {
		item.AppName = *patch.AppName
	}
	if patch.Status != nil {
		item.Status = *patch.Status
	}
	if patch.TotalQuota != nil {
		item.TotalQuota = *patch.TotalQuota
	}
	if patch.Remark != nil {
		item.Remark = *patch.Remark
	}
	if err := h.apps.Save(c.Request.Context(), item); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if patch.TotalQuota != nil {
		_ = h.svc.SyncQuotaToRedis(c.Request.Context(), item.AppID, item.TotalQuota)
	}
	response.OK(c, item)
}

func (h *AdminHandler) Delete(c *gin.Context) {
	ids, err := parseid.BindDeleteIDs(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if err := h.apps.DeleteByIDs(c.Request.Context(), ids); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *AdminHandler) RegenerateKeys(c *gin.Context) {
	id, err := parseid.GinKey(c, "id")
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	idNum, _ := parseid.Path(id)
	app, appSecret, err := h.svc.RegenerateKeys(c.Request.Context(), idNum)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, createAppResp{
		OpenApp:    *app,
		AppSecret:  appSecret,
		SignSecret: app.SignSecret,
	})
}

func (h *AdminHandler) ListDailyStat(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.stats.ListDailyStat(c.Request.Context(), store.PageQuery{Page: page, Size: size}, store.OpenAPIStatFilter{
		AppID:    c.Query("app_id"),
		Action:   c.Query("action"),
		StatDate: c.Query("stat_date"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, list, int(total), page, size)
}

func (h *AdminHandler) ListCallLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.stats.ListCallLogs(c.Request.Context(), store.PageQuery{Page: page, Size: size}, store.OpenAPIStatFilter{
		AppID:    c.Query("app_id"),
		Action:   c.Query("action"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, list, int(total), page, size)
}
