package database

import (
	"log"
	"sync"
	"time"
)

type startupProfiler struct {
	mu    sync.Mutex
	steps map[string]time.Duration
	total time.Time
}

func newStartupProfiler() *startupProfiler {
	return &startupProfiler{
		steps: make(map[string]time.Duration),
		total: time.Now(),
	}
}

func (p *startupProfiler) step(name string, start time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.steps[name] = time.Since(start)
}

func (p *startupProfiler) finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	total := time.Since(p.total)
	log.Printf("[startup] ready in %v (mysql=%v migrate=%v seed=%v redis=%v)",
		total,
		p.steps["mysql"],
		p.steps["migrate"],
		p.steps["seed"],
		p.steps["redis"],
	)
}
