package policy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/renvins/patrol/internal/engine"
	"github.com/renvins/patrol/pkg/slo"
)

type Condition struct {
	Field    string
	Operator string
	Value    float64
}

func parse(when string) (*Condition, error) {
	fields := strings.Fields(when)
	if len(fields) != 3 {
		return nil, fmt.Errorf("invalid condition %q: expected 'field operator value'", when)
	}
	percentage := strings.TrimSuffix(fields[2], "%")
	value, err := strconv.ParseFloat(percentage, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid value in condition %q: %w", when, err)
	}
	return &Condition{Field: fields[0], Operator: fields[1], Value: value}, nil
}

func evaluate(c Condition, status engine.BudgetStatus) bool {
	var actual float64
	switch c.Field {
	case "budget_remaining":
		actual = status.BudgetRemaining
	case "burn_rate":
		actual = status.BurnRate
	case "budget_consumed":
		actual = status.BudgetConsumed * 100
	default:
		slog.Warn("unknown condition field", "field", c.Field)
		return false
	}

	switch c.Operator {
	case "<":
		return actual < c.Value
	case ">":
		return actual > c.Value
	case "<=":
		return actual <= c.Value
	case ">=":
		return actual >= c.Value
	case "==":
		return actual == c.Value
	default:
		slog.Warn("unknown condition operator", "operator", c.Operator)
		return false
	}
}

type Evaluator struct {
	engine   *engine.Engine
	config   *slo.SLOConfig
	client   *http.Client
	interval time.Duration
	fired    map[string]bool
	mu       sync.Mutex
}

func New(e *engine.Engine, config *slo.SLOConfig, interval time.Duration) *Evaluator {
	return &Evaluator{
		engine:   e,
		config:   config,
		client:   &http.Client{Timeout: 10 * time.Second},
		interval: interval,
		fired:    make(map[string]bool),
	}
}

func (e *Evaluator) Run(ctx context.Context) error {
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

func (e *Evaluator) tick(ctx context.Context) {
	for _, s := range e.config.SLOs {
		status, ok := e.engine.GetStatus(s.Name)
		if !ok {
			continue
		}
		for _, p := range s.Policies {
			key := s.Name + "|" + p.When

			cond, err := parse(p.When)
			if err != nil {
				slog.Error("failed to parse policy condition", "when", p.When, "error", err)
				continue
			}

			triggered := evaluate(*cond, status)

			e.mu.Lock()
			alreadyFired := e.fired[key]
			e.mu.Unlock()

			if triggered && !alreadyFired {
				slog.Info("policy triggered", "slo", s.Name, "when", p.When, "action", p.Action)
				if err := e.fireWebhook(ctx, p.Target, status); err != nil {
					slog.Error("webhook failed", "target", p.Target, "error", err)
				} else {
					e.mu.Lock()
					e.fired[key] = true
					e.mu.Unlock()
				}
			} else if !triggered && alreadyFired {
				// condition resolved — reset so it can fire again next time
				e.mu.Lock()
				e.fired[key] = false
				e.mu.Unlock()
			}
		}
	}
}

type webhookPayload struct {
	SLOName         string  `json:"slo_name"`
	BurnRate        float64 `json:"burn_rate"`
	BudgetRemaining float64 `json:"budget_remaining"`
}

func (e *Evaluator) fireWebhook(ctx context.Context, target string, status engine.BudgetStatus) error {
	payload := webhookPayload{
		SLOName:         status.SLOName,
		BurnRate:        status.BurnRate,
		BudgetRemaining: status.BudgetRemaining,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
