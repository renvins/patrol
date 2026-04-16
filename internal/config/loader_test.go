package config

import (
	"testing"

	"github.com/renvins/patrol/pkg/slo"
)

func TestValidate(t *testing.T) {
	validSLO := slo.SLO{
		Name:      "checkout-availability",
		Service:   "checkout",
		Objective: 99.9,
		Window:    "30d",
	}

	tests := []struct {
		name    string
		cfg     *slo.SLOConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     &slo.SLOConfig{SLOs: []slo.SLO{validSLO}},
			wantErr: false,
		},
		{
			name:    "valid objective boundary 100",
			cfg:     &slo.SLOConfig{SLOs: []slo.SLO{{Name: "x", Objective: 100}}},
			wantErr: false,
		},
		{
			name:    "valid objective boundary 0.1",
			cfg:     &slo.SLOConfig{SLOs: []slo.SLO{{Name: "x", Objective: 0.1}}},
			wantErr: false,
		},
		{
			name:    "empty slos",
			cfg:     &slo.SLOConfig{SLOs: []slo.SLO{}},
			wantErr: true,
		},
		{
			name:    "slo with empty name",
			cfg:     &slo.SLOConfig{SLOs: []slo.SLO{{Name: "", Objective: 99.9}}},
			wantErr: true,
		},
		{
			name:    "objective zero",
			cfg:     &slo.SLOConfig{SLOs: []slo.SLO{{Name: "x", Objective: 0}}},
			wantErr: true,
		},
		{
			name:    "objective above 100",
			cfg:     &slo.SLOConfig{SLOs: []slo.SLO{{Name: "x", Objective: 101}}},
			wantErr: true,
		},
		{
			name:    "objective negative",
			cfg:     &slo.SLOConfig{SLOs: []slo.SLO{{Name: "x", Objective: -1}}},
			wantErr: true,
		},
		{
			name: "second slo invalid, first valid",
			cfg: &slo.SLOConfig{SLOs: []slo.SLO{
				validSLO,
				{Name: "", Objective: 99.9},
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
