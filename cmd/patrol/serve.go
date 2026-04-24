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
	"github.com/renvins/patrol/internal/grpcserver"
	"github.com/renvins/patrol/internal/policy"
	"github.com/renvins/patrol/internal/prometheus"
	"github.com/spf13/cobra"
)

var configPath string
var listenAddr string
var grpcAddr string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Patrol daemon",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVar(&configPath, "config", "patrol.yml", "path to SLO config file")
	serveCmd.Flags().StringVar(&listenAddr, "listen", ":8080", "REST API listen address")
	serveCmd.Flags().StringVar(&grpcAddr, "grpc", ":9090", "gRPC listen address")
}

func runServe(cmd *cobra.Command, args []string) error {
	sloConfig, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	prometheusClient := prometheus.New("http://localhost:9091")
	eng := engine.New(prometheusClient, sloConfig, 30*time.Second)
	restServer := api.New(eng)
	grpcServer := grpcserver.New(eng)
	policyEvaluator := policy.New(eng, sloConfig, 30*time.Second)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	engineErrChan := make(chan error, 1)
	restErrChan := make(chan error, 1)
	grpcErrChan := make(chan error, 1)
	policyErrChan := make(chan error, 1)

	go func() {
		if err := eng.Run(ctx); err != nil {
			engineErrChan <- err
		}
	}()

	go func() {
		if err := restServer.Run(ctx, listenAddr); err != nil {
			restErrChan <- err
		}
	}()

	go func() {
		if err := grpcServer.Run(ctx, grpcAddr); err != nil {
			grpcErrChan <- err
		}
	}()

	go func() {
		if err := policyEvaluator.Run(ctx); err != nil {
			policyErrChan <- err
		}
	}()

	slog.Info("patrol started", "rest", listenAddr, "grpc", grpcAddr, "config", configPath)

	select {
	case err := <-engineErrChan:
		slog.Error("engine error", "error", err)
	case err := <-restErrChan:
		slog.Error("rest server error", "error", err)
	case err := <-grpcErrChan:
		slog.Error("grpc server error", "error", err)
	case err := <-policyErrChan:
		slog.Error("policy evaluator error", "error", err)
	}
	return nil
}
