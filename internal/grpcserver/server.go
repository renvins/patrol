package grpcserver

import (
	"context"

	"github.com/renvins/patrol/internal/engine"
	patrolv1 "github.com/renvins/patrol/proto/patrol/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	patrolv1.UnimplementedPatrolServiceServer // embed this
	engine                                    *engine.Engine
}

func New(e *engine.Engine) *Server {
	return &Server{engine: e}
}

func (s *Server) GetStatus(ctx context.Context, req *patrolv1.GetStatusRequest) (*patrolv1.GetStatusResponse, error) {
	budgetStatus, ok := s.engine.GetStatus(req.SloName)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "SLO %q not found", req.SloName)
	}

}
