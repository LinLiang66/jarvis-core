package service

import (
	"context"
	"log"
)

func (e *Engine) RefreshDispatchGate(ctx context.Context) {
	// 仅启用中的任务参与轮询门控：无任务或全部停用时 needsDispatch=false。
	// 已停用任务遗留的 PENDING 实例不触发 worker 轮询。
	enabled, err := e.stores.ListEnabledJobHandlers(ctx)
	if err != nil {
		log.Printf("[scheduler] refresh dispatch gate (enabled jobs): %v", err)
		enabled = map[string]struct{}{}
	}
	online, err := e.stores.CollectOnlineWorkerHandlers(ctx)
	if err != nil {
		log.Printf("[scheduler] refresh dispatch gate (workers): %v", err)
		online = map[string]struct{}{}
	}

	target := enabled

	needs := false
	for h := range target {
		if _, ok := online[h]; ok {
			needs = true
			break
		}
	}

	e.discardInstancesWithoutWorkers(ctx, enabled, online)

	e.gateMu.Lock()
	e.needsDispatch = needs
	e.targetHandlers = target
	e.onlineHandlers = online
	e.gateMu.Unlock()
}

func (e *Engine) RefreshAndWake(ctx context.Context) {
	e.RefreshDispatchGate(ctx)
	if e.NeedsDispatch() {
		e.WakeDispatch()
	}
}

func (e *Engine) NeedsDispatch() bool {
	e.gateMu.RLock()
	defer e.gateMu.RUnlock()
	return e.needsDispatch
}

func (e *Engine) ShouldPollWorker(handlers []string) bool {
	e.gateMu.RLock()
	defer e.gateMu.RUnlock()
	if !e.needsDispatch {
		return false
	}
	for _, h := range handlers {
		if _, ok := e.targetHandlers[h]; ok {
			return true
		}
	}
	return false
}

func (e *Engine) WakeDispatch() {
	select {
	case e.dispatchWake <- struct{}{}:
	default:
	}
}
