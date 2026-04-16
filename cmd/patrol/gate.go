package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

type gateResponse struct {
	Allow  bool   `json:"allow"`
	Reason string `json:"reason"`
}

var gateCmd = &cobra.Command{
	Use:   "gate <slo-name>",
	Short: "Check whether a deployment is allowed based on error budget",
	Args:  cobra.ExactArgs(1),
	RunE:  runGate,
}

func runGate(cmd *cobra.Command, args []string) error {
	url := apiAddr + "/api/v1/gate/" + args[0]

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to reach patrol API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("SLO %q not found", args[0])
	}

	var g gateResponse
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if g.Allow {
		fmt.Printf("ALLOW — error budget healthy\n")
	} else {
		fmt.Printf("DENY  — %s\n", g.Reason)
	}
	return nil
}
