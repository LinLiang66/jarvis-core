package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"jarvis-core/scheduler/internal/config"
	"jarvis-core/scheduler/internal/model"
	redisc "jarvis-core/scheduler/internal/redis"
	"jarvis-core/scheduler/internal/store"
)

type Engine struct {
	cfg    *config.Config
	stores *store.Stores
	rdb    *redisc.Client
	cron   *cron.Cron
	mu     sync.Mutex
	entries map[int64]cron.EntryID

	gateMu         sync.RWMutex
	needsDispatch  bool
	targetHandlers map[string]struct{}
	onlineHandlers map[string]struct{}
	dispatchWake   chan struct{}
}

func NewEngine(cfg *config.Config, stores *store.Stores, rdb *redisc.Client) *Engine {
	return &Engine{
		cfg:            cfg,
		stores:         stores,
		rdb:            rdb,
		cron:           cron.New(cron.WithSeconds()),
		entries:        make(map[int64]cron.EntryID),
		targetHandlers: make(map[string]struct{}),
		onlineHandlers: make(map[string]struct{}),
		dispatchWake:   make(chan struct{}, 1),
	}
}

func (e *Engine) Start(ctx context.Context) error {
	jobs, err := e.stores.ListEnabledJobs(ctx)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		e.registerCron(job)
	}
	e.cron.Start()
	go e.workerCleanupLoop(ctx)
	go e.instanceScannerLoop(ctx)
	e.RefreshDispatchGate(ctx)
	log.Printf("[scheduler] engine started, %d cron jobs loaded, dispatch=%v", len(jobs), e.NeedsDispatch())
	return nil
}

func (e *Engine) Stop() {
	e.cron.Stop()
}

func (e *Engine) ReloadJob(ctx context.Context, job *model.JobDefinition) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if entryID, ok := e.entries[job.ID]; ok {
		e.cron.Remove(entryID)
		delete(e.entries, job.ID)
	}
	if job.Status == model.StatusEnabled && job.CronExpr != "" {
		e.registerCronLocked(*job)
	}
	e.RefreshAndWake(ctx)
}

func (e *Engine) RemoveJob(ctx context.Context, jobID int64) {
	e.mu.Lock()
	if entryID, ok := e.entries[jobID]; ok {
		e.cron.Remove(entryID)
		delete(e.entries, jobID)
	}
	e.mu.Unlock()
	e.RefreshAndWake(ctx)
}

func (e *Engine) registerCron(job model.JobDefinition) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.registerCronLocked(job)
}

func (e *Engine) registerCronLocked(job model.JobDefinition) {
	if job.CronExpr == "" {
		return
	}
	jobID := job.ID
	expr := job.CronExpr
	entryID, err := e.cron.AddFunc(expr, func() {
		e.triggerCron(context.Background(), jobID)
	})
	if err != nil {
		log.Printf("[scheduler] invalid cron job=%d expr=%s: %v", jobID, expr, err)
		return
	}
	e.entries[jobID] = entryID
}

func (e *Engine) triggerCron(ctx context.Context, jobID int64) {
	lockKey := redisc.TriggerLockKey(jobID)
	token := uuid.NewString()
	ok, err := e.rdb.SetNX(ctx, lockKey, token, 55*time.Second)
	if err != nil || !ok {
		return
	}
	defer func() { _ = e.rdb.Unlock(ctx, lockKey, token) }()

	job, err := e.stores.GetJob(ctx, jobID)
	if err != nil || job.Status != model.StatusEnabled {
		return
	}
	_, err = e.createAndDispatch(ctx, job, model.TriggerCron)
	if err != nil {
		log.Printf("[scheduler] cron trigger job=%d failed: %v", jobID, err)
	}
}

func (e *Engine) TriggerManual(ctx context.Context, jobID int64) (*model.JobInstance, error) {
	job, err := e.stores.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return e.createAndDispatch(ctx, job, model.TriggerManual)
}

func (e *Engine) createAndDispatch(ctx context.Context, job *model.JobDefinition, triggerType string) (*model.JobInstance, error) {
	inst := &model.JobInstance{
		JobID:       job.ID,
		JobName:     job.Name,
		Handler:     job.Handler,
		TriggerType: triggerType,
		Params:      job.Params,
		TimeoutSec:  job.TimeoutSec,
		Status:      model.InstPending,
	}
	if err := e.stores.CreateInstance(ctx, inst); err != nil {
		return nil, err
	}
	if err := e.dispatch(ctx, job, inst); err != nil {
		return inst, err
	}
	_ = e.stores.AppendLog(ctx, inst.ID, "info", fmt.Sprintf("instance created trigger=%s status=%s", triggerType, inst.Status))
	e.RefreshAndWake(ctx)
	return inst, nil
}

func (e *Engine) dispatch(ctx context.Context, job *model.JobDefinition, inst *model.JobInstance) error {
	switch job.BlockStrategy {
	case model.BlockSerial:
		return e.dispatchSerial(ctx, job.ID, inst)
	case model.BlockDiscard:
		return e.dispatchDiscard(ctx, job.ID, inst)
	default:
		inst.Status = model.InstPending
		return e.stores.UpdateInstance(ctx, inst)
	}
}

func (e *Engine) dispatchSerial(ctx context.Context, jobID int64, inst *model.JobInstance) error {
	lockKey := redisc.RunningLockKey(jobID)
	token := strconv.FormatInt(inst.ID, 10)
	ok, err := e.rdb.SetNX(ctx, lockKey, token, e.cfg.RunningLockTTL)
	if err != nil {
		return err
	}
	if ok {
		inst.Status = model.InstPending
	} else {
		inst.Status = model.InstQueued
		if err := e.rdb.RPush(ctx, redisc.SerialQueueKey(jobID), inst.ID); err != nil {
			return err
		}
	}
	return e.stores.UpdateInstance(ctx, inst)
}

func (e *Engine) dispatchDiscard(ctx context.Context, jobID int64, inst *model.JobInstance) error {
	exists, err := e.rdb.Exists(ctx, redisc.RunningLockKey(jobID))
	if err != nil {
		return err
	}
	if exists {
		now := time.Now()
		inst.Status = model.InstDiscarded
		inst.FinishedAt = &now
		return e.stores.UpdateInstance(ctx, inst)
	}
	token := strconv.FormatInt(inst.ID, 10)
	ok, err := e.rdb.SetNX(ctx, redisc.RunningLockKey(jobID), token, e.cfg.RunningLockTTL)
	if err != nil {
		return err
	}
	if !ok {
		now := time.Now()
		inst.Status = model.InstDiscarded
		inst.FinishedAt = &now
		return e.stores.UpdateInstance(ctx, inst)
	}
	inst.Status = model.InstPending
	return e.stores.UpdateInstance(ctx, inst)
}

func (e *Engine) OnFinish(ctx context.Context, inst *model.JobInstance) error {
	job, err := e.stores.GetJob(ctx, inst.JobID)
	if err != nil {
		return err
	}
	if job.BlockStrategy == model.BlockSerial || job.BlockStrategy == model.BlockDiscard {
		lockKey := redisc.RunningLockKey(job.ID)
		token := strconv.FormatInt(inst.ID, 10)
		_ = e.rdb.Unlock(ctx, lockKey, token)
	}
	if job.BlockStrategy == model.BlockSerial {
		if err := e.popSerialQueue(ctx, job.ID); err != nil {
			return err
		}
	}
	e.RefreshDispatchGate(ctx)
	return nil
}

func (e *Engine) popSerialQueue(ctx context.Context, jobID int64) error {
	for {
		nextIDStr, err := e.rdb.LPop(ctx, redisc.SerialQueueKey(jobID))
		if err != nil || nextIDStr == "" {
			_ = e.rdb.Del(ctx, redisc.RunningLockKey(jobID))
			return nil
		}
		nextID, err := strconv.ParseInt(nextIDStr, 10, 64)
		if err != nil {
			continue
		}
		nextInst, err := e.stores.GetInstance(ctx, nextID)
		if err != nil {
			continue
		}
		if nextInst.Status != model.InstQueued {
			continue
		}
		token := strconv.FormatInt(nextInst.ID, 10)
		ok, err := e.rdb.SetNX(ctx, redisc.RunningLockKey(jobID), token, e.cfg.RunningLockTTL)
		if err != nil || !ok {
			_ = e.rdb.RPush(ctx, redisc.SerialQueueKey(jobID), nextInst.ID)
			return nil
		}
		nextInst.Status = model.InstPending
		if err := e.stores.UpdateInstance(ctx, nextInst); err != nil {
			return err
		}
		_ = e.stores.AppendLog(ctx, nextInst.ID, "info", "dequeued from serial queue")
		return nil
	}
}

type PollTask struct {
	InstanceID int64  `json:"instance_id"`
	JobID      int64  `json:"job_id"`
	Handler    string `json:"handler"`
	Params     string `json:"params"`
	TimeoutSec int    `json:"timeout_sec"`
}

type PollResult struct {
	Task       *PollTask
	ShouldPoll bool
}

func (e *Engine) Poll(ctx context.Context, workerID string, handlers []string) (*PollResult, error) {
	if !e.ShouldPollWorker(handlers) {
		return &PollResult{ShouldPoll: false}, nil
	}
	deadline := time.Now().Add(e.cfg.PollTimeout)
	for {
		inst, err := e.stores.ClaimPendingInstance(ctx, workerID, handlers)
		if err == nil {
			timeoutSec := inst.TimeoutSec
			if timeoutSec <= 0 {
				timeoutSec = 300
			}
			_, _ = e.rdb.Incr(ctx, redisc.RoundRobinKey(inst.JobID))
			return &PollResult{
				Task: &PollTask{
					InstanceID: inst.ID,
					JobID:      inst.JobID,
					Handler:    inst.Handler,
					Params:     inst.Params,
					TimeoutSec: timeoutSec,
				},
				ShouldPoll: true,
			}, nil
		}
		if time.Now().After(deadline) {
			return &PollResult{ShouldPoll: e.ShouldPollWorker(handlers)}, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-e.dispatchWake:
			if !e.ShouldPollWorker(handlers) {
				return &PollResult{ShouldPoll: false}, nil
			}
		case <-time.After(e.cfg.PollInterval):
		}
	}
}

func (e *Engine) ReportStart(ctx context.Context, instanceID int64, workerID string) error {
	inst, err := e.stores.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if inst.WorkerID != "" && inst.WorkerID != workerID {
		return fmt.Errorf("instance assigned to another worker")
	}
	now := time.Now()
	inst.Status = model.InstRunning
	inst.WorkerID = workerID
	inst.StartedAt = &now
	if err := e.stores.UpdateInstance(ctx, inst); err != nil {
		return err
	}
	return e.stores.AppendLog(ctx, instanceID, "info", "execution started")
}

func (e *Engine) ReportFinish(ctx context.Context, instanceID int64, workerID, result string) error {
	inst, err := e.stores.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if inst.WorkerID != workerID {
		return fmt.Errorf("worker mismatch")
	}
	now := time.Now()
	inst.Status = model.InstSuccess
	inst.Result = result
	inst.FinishedAt = &now
	if err := e.stores.UpdateInstance(ctx, inst); err != nil {
		return err
	}
	_ = e.stores.AppendLog(ctx, instanceID, "info", "execution finished")
	return e.OnFinish(ctx, inst)
}

func (e *Engine) ReportFail(ctx context.Context, instanceID int64, workerID, errMsg string) error {
	inst, err := e.stores.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if inst.WorkerID != workerID {
		return fmt.Errorf("worker mismatch")
	}
	now := time.Now()
	inst.Status = model.InstFailed
	inst.ErrorMsg = errMsg
	inst.FinishedAt = &now
	if err := e.stores.UpdateInstance(ctx, inst); err != nil {
		return err
	}
	_ = e.stores.AppendLog(ctx, instanceID, "error", errMsg)
	return e.OnFinish(ctx, inst)
}

func (e *Engine) RegisterWorker(ctx context.Context, workerID, instanceID, hostname string, handlers []string) error {
	raw, _ := json.Marshal(handlers)
	now := time.Now()
	if err := e.stores.UpsertWorker(ctx, &model.WorkerNode{
		WorkerID:        workerID,
		InstanceID:      instanceID,
		Hostname:        hostname,
		Handlers:        string(raw),
		Status:          model.WorkerOnline,
		LastHeartbeatAt: &now,
	}); err != nil {
		return err
	}
	e.RefreshAndWake(ctx)
	return nil
}

func (e *Engine) Heartbeat(ctx context.Context, workerID string) error {
	return e.stores.TouchWorker(ctx, workerID, time.Now())
}

func (e *Engine) workerCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-e.cfg.WorkerTTL)
			n, _ := e.stores.MarkStaleWorkersOffline(ctx, cutoff)
			if n > 0 {
				e.RefreshDispatchGate(ctx)
			}
		}
	}
}

func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
