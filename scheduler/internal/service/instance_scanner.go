package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"jarvis-core/scheduler/internal/model"
	redisc "jarvis-core/scheduler/internal/redis"
)

const (
	logClaimTimeoutReclaimed = "认领超时，实例已回收"
	logExecutionTimeout      = "执行超时"
	logNoWorkerDiscarded     = "无在线 Worker，任务已丢弃"
)

func (e *Engine) instanceScannerLoop(ctx context.Context) {
	ticker := time.NewTicker(e.cfg.InstanceScanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.runInstanceScan(ctx)
		}
	}
}

func (e *Engine) runInstanceScan(ctx context.Context) {
	lockKey := redisc.InstanceScannerLockKey()
	token := uuid.NewString()
	ok, err := e.rdb.SetNX(ctx, lockKey, token, e.cfg.InstanceScanInterval)
	if err != nil || !ok {
		return
	}
	defer func() { _ = e.rdb.Unlock(ctx, lockKey, token) }()

	e.scanClaimTimeouts(ctx)
	e.scanRunningTimeouts(ctx)

	enabled, err := e.stores.ListEnabledJobHandlers(ctx)
	if err != nil {
		log.Printf("[scheduler] instance scan (enabled jobs): %v", err)
		return
	}
	online, err := e.stores.CollectOnlineWorkerHandlers(ctx)
	if err != nil {
		log.Printf("[scheduler] instance scan (workers): %v", err)
		return
	}
	e.discardInstancesWithoutWorkers(ctx, enabled, online)
}

func (e *Engine) scanClaimTimeouts(ctx context.Context) {
	cutoff := time.Now().Add(-e.cfg.InstanceClaimTimeout)
	list, err := e.stores.ListStaleClaimedPending(ctx, cutoff)
	if err != nil {
		log.Printf("[scheduler] scan claim timeout: %v", err)
		return
	}
	for _, inst := range list {
		reclaimed, err := e.stores.ReclaimInstance(ctx, inst.ID, inst.WorkerID)
		if err != nil {
			log.Printf("[scheduler] reclaim instance=%d: %v", inst.ID, err)
			continue
		}
		if reclaimed {
			_ = e.stores.AppendLog(ctx, inst.ID, "warn", logClaimTimeoutReclaimed)
			log.Printf("[scheduler] instance=%d claim timeout, reclaimed from worker=%s", inst.ID, inst.WorkerID)
		}
	}
}

func (e *Engine) scanRunningTimeouts(ctx context.Context) {
	now := time.Now()
	list, err := e.stores.ListTimedOutRunning(ctx, now)
	if err != nil {
		log.Printf("[scheduler] scan running timeout: %v", err)
		return
	}
	for _, inst := range list {
		if err := e.failTimedOutInstance(ctx, &inst); err != nil {
			log.Printf("[scheduler] fail timed out instance=%d: %v", inst.ID, err)
		}
	}
}

func (e *Engine) failTimedOutInstance(ctx context.Context, inst *model.JobInstance) error {
	fresh, err := e.stores.GetInstance(ctx, inst.ID)
	if err != nil {
		return err
	}
	if fresh.Status != model.InstRunning {
		return nil
	}
	now := time.Now()
	fresh.Status = model.InstFailed
	fresh.ErrorMsg = logExecutionTimeout
	fresh.FinishedAt = &now
	if err := e.stores.UpdateInstance(ctx, fresh); err != nil {
		return err
	}
	_ = e.stores.AppendLog(ctx, fresh.ID, "error", logExecutionTimeout)
	log.Printf("[scheduler] instance=%d execution timeout", fresh.ID)
	return e.OnFinish(ctx, fresh)
}

func (e *Engine) discardInstancesWithoutWorkers(ctx context.Context, enabled, online map[string]struct{}) {
	noWorkerHandlers := handlersWithoutOnlineWorkers(enabled, online)
	if len(noWorkerHandlers) == 0 {
		return
	}
	const maxRounds = 100
	for round := 0; round < maxRounds; round++ {
		instances, err := e.stores.ListUndispatchedByHandlers(ctx, noWorkerHandlers)
		if err != nil {
			log.Printf("[scheduler] list undispatched instances: %v", err)
			return
		}
		if len(instances) == 0 {
			return
		}
		discarded := 0
		for _, inst := range instances {
			if err := e.discardInstance(ctx, inst.ID, logNoWorkerDiscarded); err != nil {
				log.Printf("[scheduler] discard instance=%d: %v", inst.ID, err)
				continue
			}
			discarded++
		}
		if discarded == 0 {
			return
		}
	}
}

func (e *Engine) discardInstance(ctx context.Context, instanceID int64, reason string) error {
	fresh, err := e.stores.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if fresh.Status != model.InstPending && fresh.Status != model.InstQueued {
		return nil
	}
	if fresh.Status == model.InstPending && fresh.WorkerID != "" {
		return nil
	}
	now := time.Now()
	fresh.Status = model.InstDiscarded
	fresh.FinishedAt = &now
	if err := e.stores.UpdateInstance(ctx, fresh); err != nil {
		return err
	}
	_ = e.stores.AppendLog(ctx, fresh.ID, "warn", reason)
	log.Printf("[scheduler] instance=%d discarded: %s", fresh.ID, reason)
	return e.OnFinish(ctx, fresh)
}

func handlersWithoutOnlineWorkers(enabled, online map[string]struct{}) []string {
	var list []string
	for h := range enabled {
		if _, ok := online[h]; !ok {
			list = append(list, h)
		}
	}
	return list
}
