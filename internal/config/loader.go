package config

import (
	"fmt"
	"os"

	"github.com/renvins/patrol/pkg/slo"
	"gopkg.in/yaml.v3"
)

func Load(path string) (*slo.SLOConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config slo.SLOConfig
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}
	err = validate(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func validate(cfg *slo.SLOConfig) error {
	if len(cfg.SLOs) == 0 {
		return fmt.Errorf("empty configuration")
	}
	for _, slo := range cfg.SLOs {
		if len(slo.Name) == 0 {
			return fmt.Errorf("slo without name was detected")
		}
		if slo.Objective <= 0 || slo.Objective > 100 {
			return fmt.Errorf("slo %q has objective outside range (0, 100]", slo.Name)
		}
	}
	return nil
}
