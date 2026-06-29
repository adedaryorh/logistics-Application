package http

import (
	"context"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/clients/healthgrpc"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	"google.golang.org/grpc"
)

func (a *App) RegisterGRPC(server *grpc.Server, cfg *platformconfig.Config) {
	healthgrpc.RegisterServer(server, &operationsHealthServer{service: "operations-service", version: cfg.Server.Version})
}

type operationsHealthServer struct {
	service string
	version string
}

func (s *operationsHealthServer) Check(ctx context.Context, _ *healthgrpc.CheckRequest) (*healthgrpc.CheckResponse, error) {
	return &healthgrpc.CheckResponse{
		Service:   s.service,
		Status:    "ok",
		Version:   s.version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
