package grpcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/renvins/patrol/internal/engine"
	patrolv1 "github.com/renvins/patrol/proto/patrol/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	patrolv1.UnimplementedPatrolServiceServer
	engine *engine.Engine
}

func New(e *engine.Engine) *Server {
	return &Server{engine: e}
}

func (s *Server) GetStatus(_ context.Context, req *patrolv1.GetStatusRequest) (*patrolv1.GetStatusResponse, error) {
	bs, ok := s.engine.GetStatus(req.SloName)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "SLO %q not found", req.SloName)
	}
	return &patrolv1.GetStatusResponse{
		SloName:         bs.SLOName,
		BurnRate:        bs.BurnRate,
		BudgetConsumed:  bs.BudgetConsumed,
		BudgetRemaining: bs.BudgetRemaining,
		ErrorRate:       bs.ErrorRate,
		UpdatedAt:       bs.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Server) GetGate(_ context.Context, req *patrolv1.GetGateRequest) (*patrolv1.GetGateResponse, error) {
	bs, ok := s.engine.GetStatus(req.SloName)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "SLO %q not found", req.SloName)
	}
	if bs.BudgetRemaining > 0 {
		return &patrolv1.GetGateResponse{Allow: true}, nil
	}
	return &patrolv1.GetGateResponse{Allow: false, Reason: "error budget exhausted"}, nil
}

func (s *Server) Run(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	grpcSrv := grpc.NewServer()
	patrolv1.RegisterPatrolServiceServer(grpcSrv, s)

	errChan := make(chan error, 1)
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("grpc server error: %w", err)
	case <-ctx.Done():
		grpcSrv.GracefulStop()
		return nil
	}
}
