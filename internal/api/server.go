package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/renvins/patrol/internal/engine"
)

type Server struct {
	engine *engine.Engine
	mux    *http.ServeMux
}

func New(e *engine.Engine) *Server {
	s := &Server{engine: e, mux: http.NewServeMux()}
	s.mux.HandleFunc("/api/v1/status/", s.handleStatus)
	s.mux.HandleFunc("/api/v1/gate/", s.handleGate)
	return s
}

func (s *Server) Run(ctx context.Context, addr string) error {
	server := http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	serverErr := make(chan error)

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		return nil
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/status/")
	if name == "" {
		http.Error(w, "missing slo name", http.StatusBadRequest)
		return
	}

	status, ok := s.engine.GetStatus(name)
	if !ok {
		http.Error(w, "slo not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleGate(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/gate/")
	if name == "" {
		http.Error(w, "missing slo name", http.StatusBadRequest)
		return
	}

	status, ok := s.engine.GetStatus(name)
	if !ok {
		http.Error(w, "slo not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if status.BudgetRemaining > 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"allow": true})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"allow":  false,
			"reason": "error budget exhausted",
		})
	}
}
