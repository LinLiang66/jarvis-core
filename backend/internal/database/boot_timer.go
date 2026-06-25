package database

import (
	"log"
	"sync"
	"time"
)

type bootTimer struct {
	start time.Time
	mu    sync.Mutex
	last  time.Time
	steps []bootStep
}

type bootStep struct {
	name string
	ms   int64
}

func newBootTimer() *bootTimer {
	now := time.Now()
	return &bootTimer{start: now, last: now}
}

func (t *bootTimer) mark(name string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	stepMS := now.Sub(t.last).Milliseconds()
	t.steps = append(t.steps, bootStep{name: name, ms: stepMS})
	t.last = now
}

func (t *bootTimer) summary() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	total := time.Since(t.start).Milliseconds()
	var detail string
	for i, s := range t.steps {
		if i > 0 {
			detail += ", "
		}
		detail += s.name + "=" + formatBootMS(s.ms)
	}
	log.Printf("[boot] startup finished total=%s (%s)", formatBootMS(total), detail)
}

func formatBootMS(ms int64) string {
	if ms < 1000 {
		return time.Duration(ms * int64(time.Millisecond)).String()
	}
	return time.Duration(ms * int64(time.Millisecond)).Round(time.Millisecond).String()
}
