package worker

import (
	"github.com/gin-gonic/gin"

	"jarvis-core/scheduler/internal/pkg/response"
	"jarvis-core/scheduler/internal/service"
)

type Handler struct {
	engine *service.Engine
}

func New(engine *service.Engine) *Handler {
	return &Handler{engine: engine}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.POST("/register", h.RegisterWorker)
	rg.POST("/heartbeat", h.Heartbeat)
	rg.POST("/ready", h.Ready)
	rg.GET("/poll", h.Poll)
	rg.POST("/poll", h.Poll)
	rg.POST("/report/start", h.ReportStart)
	rg.POST("/report/finish", h.ReportFinish)
	rg.POST("/report/fail", h.ReportFail)
}

type registerReq struct {
	WorkerID   string   `json:"worker_id" binding:"required"`
	InstanceID string   `json:"instance_id"`
	Hostname   string   `json:"hostname"`
	Handlers   []string `json:"handlers" binding:"required"`
}

func (h *Handler) RegisterWorker(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.engine.RegisterWorker(c.Request.Context(), req.WorkerID, req.InstanceID, req.Hostname, req.Handlers); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"should_poll": h.engine.ShouldPollWorker(req.Handlers)})
}

type workerIDReq struct {
	WorkerID string `json:"worker_id" binding:"required"`
}

func (h *Handler) Heartbeat(c *gin.Context) {
	var req workerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.engine.Heartbeat(c.Request.Context(), req.WorkerID); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}

type readyReq struct {
	Handlers []string `json:"handlers" binding:"required"`
}

func (h *Handler) Ready(c *gin.Context) {
	var req readyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	response.OK(c, gin.H{"should_poll": h.engine.ShouldPollWorker(req.Handlers)})
}

type pollReq struct {
	WorkerID string   `json:"worker_id" binding:"required"`
	Handlers []string `json:"handlers" binding:"required"`
}

type pollRespData struct {
	ShouldPoll bool `json:"should_poll"`
	service.PollTask
}

func (h *Handler) Poll(c *gin.Context) {
	var req pollReq
	if c.Request.Method == "GET" {
		req.WorkerID = c.Query("worker_id")
		req.Handlers = c.QueryArray("handlers")
	} else if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.WorkerID == "" || len(req.Handlers) == 0 {
		response.Fail(c, 400, "worker_id 与 handlers 必填")
		return
	}
	result, err := h.engine.Poll(c.Request.Context(), req.WorkerID, req.Handlers)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	data := pollRespData{ShouldPoll: result.ShouldPoll}
	if result.Task != nil {
		data.PollTask = *result.Task
	}
	response.OK(c, data)
}

type reportStartReq struct {
	WorkerID   string `json:"worker_id" binding:"required"`
	InstanceID int64  `json:"instance_id" binding:"required"`
}

func (h *Handler) ReportStart(c *gin.Context) {
	var req reportStartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.engine.ReportStart(c.Request.Context(), req.InstanceID, req.WorkerID); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}

type reportFinishReq struct {
	WorkerID   string `json:"worker_id" binding:"required"`
	InstanceID int64  `json:"instance_id" binding:"required"`
	Result     string `json:"result"`
}

func (h *Handler) ReportFinish(c *gin.Context) {
	var req reportFinishReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.engine.ReportFinish(c.Request.Context(), req.InstanceID, req.WorkerID, req.Result); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}

type reportFailReq struct {
	WorkerID   string `json:"worker_id" binding:"required"`
	InstanceID int64  `json:"instance_id" binding:"required"`
	Error      string `json:"error"`
}

func (h *Handler) ReportFail(c *gin.Context) {
	var req reportFailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.engine.ReportFail(c.Request.Context(), req.InstanceID, req.WorkerID, req.Error); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}
