package slo

import (
	"math"
	"testing"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestBurnRate(t *testing.T) {
	tests := []struct {
		name      string
		errorRate float64
		objective float64
		want      float64
	}{
		{
			name:      "zero errors",
			errorRate: 0,
			objective: 99.9,
			want:      0,
		},
		{
			name:      "burn rate exactly 1 (sustainable)",
			errorRate: 0.001,
			objective: 99.9,
			want:      1.0,
		},
		{
			name:      "burn rate 14.4 (critical threshold)",
			errorRate: 0.0144,
			objective: 99.9,
			want:      14.4,
		},
		{
			name:      "healthy service, burn rate below 1",
			errorRate: 0.0001,
			objective: 99.9,
			want:      0.1,
		},
		{
			name:      "99.0 objective",
			errorRate: 0.01,
			objective: 99.0,
			want:      1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BurnRate(tt.errorRate, tt.objective)
			if !approxEqual(got, tt.want) {
				t.Errorf("BurnRate(%v, %v) = %v, want %v", tt.errorRate, tt.objective, got, tt.want)
			}
		})
	}
}

func TestBudgetConsumed(t *testing.T) {
	tests := []struct {
		name      string
		errorRate float64
		objective float64
		want      float64
	}{
		{
			name:      "zero errors",
			errorRate: 0,
			objective: 99.9,
			want:      0,
		},
		{
			name:      "exactly at budget",
			errorRate: 0.001,
			objective: 99.9,
			want:      1.0,
		},
		{
			name:      "budget exceeded",
			errorRate: 0.0144,
			objective: 99.9,
			want:      14.4,
		},
		{
			name:      "10% of budget consumed",
			errorRate: 0.0001,
			objective: 99.9,
			want:      0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BudgetConsumed(tt.errorRate, tt.objective)
			if !approxEqual(got, tt.want) {
				t.Errorf("BudgetConsumed(%v, %v) = %v, want %v", tt.errorRate, tt.objective, got, tt.want)
			}
		})
	}
}

func TestBudgetRemaining(t *testing.T) {
	tests := []struct {
		name      string
		errorRate float64
		objective float64
		want      float64
	}{
		{
			name:      "zero errors, 100% remaining",
			errorRate: 0,
			objective: 99.9,
			want:      100.0,
		},
		{
			name:      "exactly at budget, 0% remaining",
			errorRate: 0.001,
			objective: 99.9,
			want:      0.0,
		},
		{
			name:      "10% consumed, 90% remaining",
			errorRate: 0.0001,
			objective: 99.9,
			want:      90.0,
		},
		{
			name:      "budget exceeded, negative remaining",
			errorRate: 0.0144,
			objective: 99.9,
			want:      -1340.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BudgetRemaining(tt.errorRate, tt.objective)
			if !approxEqual(got, tt.want) {
				t.Errorf("BudgetRemaining(%v, %v) = %v, want %v", tt.errorRate, tt.objective, got, tt.want)
			}
		})
	}
}
