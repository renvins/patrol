package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/renvins/patrol/internal/api"
	"github.com/renvins/patrol/internal/config"
	"github.com/renvins/patrol/internal/engine"
	"github.com/renvins/patrol/internal/prometheus"
	"github.com/spf13/cobra"
)

var configPath string
var listenAddr string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Patrol daemon",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVar(&configPath, "config", "patrol.yml", "path to SLO config file")
	serveCmd.Flags().StringVar(&listenAddr, "listen", ":8080", "address to listen on")
}

func runServe(cmd *cobra.Command, args []string) error {
	sloConfig, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	prometheusClient := prometheus.New("http://localhost:9090")
	eng := engine.New(prometheusClient, sloConfig, 30*time.Second)
	server := api.New(eng)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	engineErrChan := make(chan error, 1)
	serverErrChan := make(chan error, 1)

	go func() {
		if err := eng.Run(ctx); err != nil {
			engineErrChan <- err
		}
	}()

	go func() {
		if err := server.Run(ctx, listenAddr); err != nil {
			serverErrChan <- err
		}
	}()

	slog.Info("patrol started", "listen", listenAddr, "config", configPath)

	select {
	case err := <-engineErrChan:
		slog.Error("engine error", "error", err)
	case err := <-serverErrChan:
		slog.Error("server error", "error", err)
	}
	return nil
}
