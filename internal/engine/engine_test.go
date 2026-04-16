package engine

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/renvins/patrol/pkg/slo"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// fakeQuerier is a fake implementation of MetricQuerier for testing.
// It returns a fixed error rate per query string, no real HTTP involved.
type fakeQuerier struct {
	results map[string]float64
}

func (f *fakeQuerier) QueryInstant(_ context.Context, query string) (float64, error) {
	return f.results[query], nil
}

var testConfig = &slo.SLOConfig{
	SLOs: []slo.SLO{
		{
			Name:      "checkout-availability",
			Service:   "checkout",
			Objective: 99.9,
			Window:    "30d",
			SLI: slo.SLI{
				Source: "prometheus",
				Query:  "error_rate_query",
			},
		},
	},
}

func TestEngine_GetStatus_BeforeTick(t *testing.T) {
	querier := &fakeQuerier{results: map[string]float64{}}
	e := New(querier, testConfig, time.Minute)

	_, ok := e.GetStatus("checkout-availability")
	if ok {
		t.Error("expected no status before first tick, got one")
	}
}

func TestEngine_Tick_PopulatesStatus(t *testing.T) {
	querier := &fakeQuerier{
		results: map[string]float64{
			"error_rate_query": 0.001, // burn rate = 1.0 exactly
		},
	}
	e := New(querier, testConfig, time.Minute)
	e.tick(context.Background())

	status, ok := e.GetStatus("checkout-availability")
	if !ok {
		t.Fatal("expected status after tick, got none")
	}
	if status.SLOName != "checkout-availability" {
		t.Errorf("SLOName = %q, want %q", status.SLOName, "checkout-availability")
	}
	if status.ErrorRate != 0.001 {
		t.Errorf("ErrorRate = %v, want 0.001", status.ErrorRate)
	}
	if !approxEqual(status.BurnRate, 1.0) {
		t.Errorf("BurnRate = %v, want 1.0", status.BurnRate)
	}
	if !approxEqual(status.BudgetRemaining, 0.0) {
		t.Errorf("BudgetRemaining = %v, want 0.0", status.BudgetRemaining)
	}
	if status.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestEngine_Tick_HighBurnRate(t *testing.T) {
	querier := &fakeQuerier{
		results: map[string]float64{
			"error_rate_query": 0.0144, // burn rate = 14.4
		},
	}
	e := New(querier, testConfig, time.Minute)
	e.tick(context.Background())

	status, ok := e.GetStatus("checkout-availability")
	if !ok {
		t.Fatal("expected status after tick")
	}
	if status.BurnRate < 14.0 {
		t.Errorf("expected high burn rate, got %v", status.BurnRate)
	}
	if status.BudgetRemaining >= 0 {
		t.Errorf("expected negative budget remaining, got %v", status.BudgetRemaining)
	}
}

func TestEngine_Tick_UpdatesStatus(t *testing.T) {
	querier := &fakeQuerier{
		results: map[string]float64{"error_rate_query": 0.0001},
	}
	e := New(querier, testConfig, time.Minute)

	e.tick(context.Background())
	first, _ := e.GetStatus("checkout-availability")

	querier.results["error_rate_query"] = 0.0005
	e.tick(context.Background())
	second, _ := e.GetStatus("checkout-availability")

	if second.ErrorRate == first.ErrorRate {
		t.Error("expected status to update after second tick")
	}
	if !second.UpdatedAt.After(first.UpdatedAt) && second.UpdatedAt.Equal(first.UpdatedAt) {
		t.Error("expected UpdatedAt to advance after second tick")
	}
}

func TestEngine_Run_StopsOnContextCancel(t *testing.T) {
	querier := &fakeQuerier{results: map[string]float64{}}
	e := New(querier, testConfig, time.Hour) // long interval so ticker never fires

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- e.Run(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("Run() returned %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Error("Run() did not stop after context was cancelled")
	}
}
