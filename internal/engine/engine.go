package engine

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/renvins/patrol/pkg/slo"
)

type MetricQuerier interface {
	QueryInstant(ctx context.Context, query string) (float64, error)
}

// BudgetStatus is the computed state for one SLO after a tick.
type BudgetStatus struct {
	SLOName         string
	BurnRate        float64
	BudgetConsumed  float64
	BudgetRemaining float64
	ErrorRate       float64
	UpdatedAt       time.Time
}

type Engine struct {
	querier   MetricQuerier
	config    *slo.SLOConfig
	statuses  map[string]BudgetStatus
	mu        sync.RWMutex
	interval  time.Duration
}

func New(querier MetricQuerier, config *slo.SLOConfig, interval time.Duration) *Engine {
	return &Engine{
		querier:  querier,
		config:   config,
		statuses: make(map[string]BudgetStatus),
		interval: interval,
	}
}

// GetStatus returns the latest computed budget status for an SLO by name.
// The second return value is false if no status exists yet for that name.
func (e *Engine) GetStatus(name string) (BudgetStatus, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	s, ok := e.statuses[name]
	return s, ok
}

// Run starts the engine loop. It blocks until ctx is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.tick(ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (e *Engine) tick(ctx context.Context) {
	for _, s := range e.config.SLOs {
		errorRate, err := e.querier.QueryInstant(ctx, s.SLI.Query)
		if err != nil {
			slog.Error("failed to query metric", "slo", s.Name, "error", err)
			continue
		}

		status := BudgetStatus{
			SLOName:         s.Name,
			ErrorRate:       errorRate,
			BurnRate:        slo.BurnRate(errorRate, s.Objective),
			BudgetConsumed:  slo.BudgetConsumed(errorRate, s.Objective),
			BudgetRemaining: slo.BudgetRemaining(errorRate, s.Objective),
			UpdatedAt:       time.Now(),
		}

		e.mu.Lock()
		e.statuses[s.Name] = status
		e.mu.Unlock()
	}
}
