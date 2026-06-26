package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"jarvis-core/scheduler/internal/model"
	"jarvis-core/scheduler/internal/pkg/response"
	"jarvis-core/scheduler/internal/service"
	"jarvis-core/scheduler/internal/store"
)

type Handler struct {
	stores *store.Stores
	engine *service.Engine
}

func New(stores *store.Stores, engine *service.Engine) *Handler {
	return &Handler{stores: stores, engine: engine}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.GET("/jobs", h.ListJobs)
	rg.GET("/jobs/:id", h.GetJob)
	rg.POST("/jobs", h.CreateJob)
	rg.PUT("/jobs/:id", h.UpdateJob)
	rg.DELETE("/jobs/:id", h.DeleteJob)
	rg.POST("/jobs/:id/trigger", h.TriggerJob)
	rg.GET("/instances", h.ListInstances)
	rg.GET("/instances/:id", h.GetInstance)
	rg.GET("/instances/:id/logs", h.ListLogs)
	rg.GET("/workers", h.ListWorkers)
}

type jobReq struct {
	GroupName     string `json:"group_name"`
	Name          string `json:"name" binding:"required"`
	Handler       string `json:"handler" binding:"required"`
	TriggerType   string `json:"trigger_type"`
	CronExpr      string `json:"cron_expr"`
	Params        string `json:"params"`
	BlockStrategy string `json:"block_strategy"`
	RouteStrategy string `json:"route_strategy"`
	ExecuteMode   string `json:"execute_mode"`
	Status        string `json:"status"`
	Description   string `json:"description"`
	TimeoutSec    int    `json:"timeout_sec"`
	RetryCount    int    `json:"retry_count"`
	RetryInterval int    `json:"retry_interval"`
	ParallelCount int    `json:"parallel_count"`
}

func (h *Handler) ListJobs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.stores.ListJobs(c.Request.Context(), store.PageQuery{Page: page, Size: size}, store.JobFilter{
		Name:    c.Query("name"),
		Handler: c.Query("handler"),
		Status:  c.Query("status"),
	})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, list, int(total), page, size)
}

func (h *Handler) GetJob(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	job, err := h.stores.GetJob(c.Request.Context(), id)
	if service.IsNotFound(err) {
		response.Fail(c, 404, "任务不存在")
		return
	}
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, job)
}

func (h *Handler) CreateJob(c *gin.Context) {
	var req jobReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	job := buildJobFromReq(req)
	if err := h.stores.CreateJob(c.Request.Context(), &job); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	h.engine.ReloadJob(c.Request.Context(), &job)
	response.OK(c, job)
}

func (h *Handler) UpdateJob(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	job, err := h.stores.GetJob(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 404, "任务不存在")
		return
	}
	var req jobReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	applyJobReq(job, req)
	if err := h.stores.UpdateJob(c.Request.Context(), job); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	h.engine.ReloadJob(c.Request.Context(), job)
	response.OK(c, job)
}

func (h *Handler) DeleteJob(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if err := h.stores.DeleteJobs(c.Request.Context(), []int64{id}); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	h.engine.RemoveJob(c.Request.Context(), id)
	response.OK(c, nil)
}

func (h *Handler) TriggerJob(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	inst, err := h.engine.TriggerManual(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, inst)
}

func (h *Handler) ListInstances(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	jobID, _ := strconv.ParseInt(c.Query("job_id"), 10, 64)
	list, total, err := h.stores.ListInstances(c.Request.Context(), store.PageQuery{Page: page, Size: size}, store.InstanceFilter{
		JobID:   jobID,
		Handler: c.Query("handler"),
		Status:  c.Query("status"),
	})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, list, int(total), page, size)
}

func (h *Handler) GetInstance(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	inst, err := h.stores.GetInstance(c.Request.Context(), id)
	if service.IsNotFound(err) {
		response.Fail(c, 404, "实例不存在")
		return
	}
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, inst)
}

func (h *Handler) ListLogs(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))
	list, total, err := h.stores.ListLogs(c.Request.Context(), id, store.PageQuery{Page: page, Size: size})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, list, int(total), page, size)
}

func (h *Handler) ListWorkers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.stores.ListWorkers(c.Request.Context(), store.PageQuery{Page: page, Size: size})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, list, int(total), page, size)
}

func buildJobFromReq(req jobReq) model.JobDefinition {
	job := model.JobDefinition{
		GroupName:     req.GroupName,
		Name:          req.Name,
		Handler:       req.Handler,
		TriggerType:   req.TriggerType,
		CronExpr:      req.CronExpr,
		Params:        req.Params,
		BlockStrategy: req.BlockStrategy,
		RouteStrategy: req.RouteStrategy,
		ExecuteMode:   req.ExecuteMode,
		Status:        req.Status,
		Description:   req.Description,
		TimeoutSec:    req.TimeoutSec,
		RetryCount:    req.RetryCount,
		RetryInterval: req.RetryInterval,
		ParallelCount: req.ParallelCount,
	}
	applyDefaults(&job)
	return job
}

func applyJobReq(job *model.JobDefinition, req jobReq) {
	job.GroupName = req.GroupName
	job.Name = req.Name
	job.Handler = req.Handler
	if req.TriggerType != "" {
		job.TriggerType = req.TriggerType
	}
	job.CronExpr = req.CronExpr
	job.Params = req.Params
	if req.BlockStrategy != "" {
		job.BlockStrategy = req.BlockStrategy
	}
	if req.RouteStrategy != "" {
		job.RouteStrategy = req.RouteStrategy
	}
	if req.ExecuteMode != "" {
		job.ExecuteMode = req.ExecuteMode
	}
	if req.Status != "" {
		job.Status = req.Status
	}
	job.Description = req.Description
	if req.TimeoutSec > 0 {
		job.TimeoutSec = req.TimeoutSec
	}
	if req.RetryCount >= 0 {
		job.RetryCount = req.RetryCount
	}
	if req.RetryInterval > 0 {
		job.RetryInterval = req.RetryInterval
	}
	if req.ParallelCount > 0 {
		job.ParallelCount = req.ParallelCount
	}
	applyDefaults(job)
}

func applyDefaults(job *model.JobDefinition) {
	if job.GroupName == "" {
		job.GroupName = "default"
	}
	if job.TriggerType == "" {
		job.TriggerType = model.TriggerCron
	}
	if job.BlockStrategy == "" {
		job.BlockStrategy = model.BlockParallel
	}
	if job.RouteStrategy == "" {
		job.RouteStrategy = model.RouteRoundRobin
	}
	if job.ExecuteMode == "" {
		job.ExecuteMode = model.ExecuteCluster
	}
	if job.Status == "" {
		job.Status = model.StatusEnabled
	}
	if job.TimeoutSec <= 0 {
		job.TimeoutSec = 300
	}
	if job.RetryInterval <= 0 {
		job.RetryInterval = 60
	}
	if job.ParallelCount <= 0 {
		job.ParallelCount = 1
	}
}

func parseID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
