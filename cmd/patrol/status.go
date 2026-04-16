package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// statusResponse mirrors engine.BudgetStatus for JSON decoding.
// We redeclare it here so the CLI doesn't import the internal engine package.
type statusResponse struct {
	SLOName         string    `json:"SLOName"`
	BurnRate        float64   `json:"BurnRate"`
	BudgetConsumed  float64   `json:"BudgetConsumed"`
	BudgetRemaining float64   `json:"BudgetRemaining"`
	ErrorRate       float64   `json:"ErrorRate"`
	UpdatedAt       time.Time `json:"UpdatedAt"`
}

var statusCmd = &cobra.Command{
	Use:   "status <slo-name>",
	Short: "Show the current budget status for an SLO",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	url := apiAddr + "/api/v1/status/" + args[0]

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to reach patrol API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("SLO %q not found", args[0])
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var s statusResponse
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("SLO:              %s\n", s.SLOName)
	fmt.Printf("Error Rate:       %.4f%%\n", s.ErrorRate*100)
	fmt.Printf("Burn Rate:        %.2fx\n", s.BurnRate)
	fmt.Printf("Budget Consumed:  %.2f%%\n", s.BudgetConsumed*100)
	fmt.Printf("Budget Remaining: %.2f%%\n", s.BudgetRemaining)
	fmt.Printf("Last Updated:     %s\n", s.UpdatedAt.Format(time.RFC3339))
	return nil
}
