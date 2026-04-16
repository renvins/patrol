package main

import "github.com/spf13/cobra"

var apiAddr string

var rootCmd = &cobra.Command{
	Use:   "patrol",
	Short: "Patrol - SLO engine and deployment gatekeeper",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiAddr, "api", "http://localhost:8080", "Patrol API address")
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(gateCmd)
	rootCmd.AddCommand(serveCmd)
}
